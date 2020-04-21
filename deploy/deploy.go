package deploy

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/xerrors"

	"github.com/ORG_NAME/REPO_NAME/server/tools/pubsub_generator/misc"
	"github.com/otiai10/copy"
)

func Main() {
	var (
		dryRun  = flag.Bool("dry-run", false, "Don't create actual resources(print the diff and exit)")
		runtime = flag.String("runtime", "go111", "Runtime for Cloud Functions")
		region  = flag.String("region", "asia-northeast1", "Region to deploy functions")
		options = flag.String("options", "", "Options for gcloud functions deploy")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if err := runDeployment(ctx, *dryRun, *runtime, *region, *options); err != nil {
		log.Fatal(err.Error())
	}
}

// prepareVendor - vendorにパッケージ自身全体を入れることで元のパッケージ名(github.com/...)で解決できるようにする
func prepareVendor(goRoot, packageName string) (string, error) {
	vendorDir, err := filepath.Abs(filepath.Join(goRoot, "vendor"))

	if err != nil {
		return "", xerrors.Errorf("failed to calc vendor dir absolute path: %w", err)
	}

	if _, err = os.Stat(vendorDir); err != nil {
		// vendor does not exist
		log.Println("Running go mod vendor")

		if _, err = exec.Command("go", "mod", "vendor").CombinedOutput(); err != nil {
			return "", xerrors.Errorf("go mod vendor failed: %w", err)
		}
	}

	tmpVendor, err := ioutil.TempDir("", "pubsub_generator_deploy_vendor")

	if err != nil {
		return "", xerrors.Errorf("failed to ")
	}

	if err := copy.Copy(vendorDir, tmpVendor); err != nil {
		return "", xerrors.Errorf("failed to copy vendor to temporary dir: %w", err)
	}

	selfVendor := filepath.Join(tmpVendor, packageName)

	if err := os.RemoveAll(selfVendor); err != nil {
		return "", xerrors.Errorf("failed to remove old directories(%s): %w", selfVendor, err)
	}

	if err := os.MkdirAll(selfVendor, 0770); err != nil {
		return "", xerrors.Errorf("failed to mkdir all for package(%s): %w", packageName, err)
	}

	if err := copy.Copy(goRoot, selfVendor, copy.Options{
		Skip: func(src string) bool {
			if vendorDir == src {
				return true
			}

			return false
		},
	}); err != nil {
		return "", xerrors.Errorf("failed to copy myself to vendor(%s): %w", selfVendor, err)
	}

	return tmpVendor, nil
}

func runDeployment(ctx context.Context, dryRun bool, runtime, region, options string) error {
	goMod := misc.GetGoModPath()

	if goMod == "." {
		return xerrors.Errorf("go package directory is not found")
	}

	goRoot, err := filepath.Abs(filepath.Dir(goMod))

	if err != nil {
		return xerrors.Errorf("failed to calc abs path: %w", err)
	}

	packageName, err := misc.GetGoRootPackageName()

	if err != nil {
		return xerrors.Errorf("failed to get package name from go.mod: %w", err)
	}

	functionsRoot := filepath.Join(goRoot, "infra/functions")
	pubsubRoot := filepath.Join(goRoot, "infra/pubsub")

	fifos, err := ioutil.ReadDir(functionsRoot)

	if err != nil {
		return err
	}

	topics := map[string]string{}
	for i := range fifos {
		if !fifos[i].IsDir() {
			continue
		}

		//nolint:govet
		topic, err := misc.FindTopicConst(filepath.Join(pubsubRoot, fifos[i].Name()))

		if err != nil {
			log.Printf("error occurred in infar/pubsub/%s: %+v", fifos[i].Name(), err)
		}

		topics[topic] = filepath.Join(functionsRoot, fifos[i].Name())
	}

	var tmpVendor string
	if tmpVendor, err = prepareVendor(goRoot, packageName); err != nil {
		return xerrors.Errorf("fail to prepare vendor/: %+v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpVendor); err != nil {
			log.Printf("vendor cleanup failed: %+v", err)
		}
	}()

	command := template.Must(
		template.
			New("command").
			Parse("gcloud functions deploy " +
				"{{.FunctionName}} " +
				"--region {{.Region}} " +
				"--runtime {{.Runtime}} " +
				"--entry-point {{.EntryPoint}} " +
				"--source {{.Source}} " +
				"--trigger-topic {{.Topic}} " +
				"{{.Options}}"),
	)

	for topic, funcDir := range topics {
		if err := func() error {
			dir, err := ioutil.TempDir("", "pubsub_generator_deploy_")

			if err != nil {
				return xerrors.Errorf("failed to generate temporary directory: %w", err)
			}

			if err := copy.Copy(funcDir, dir); err != nil {
				return xerrors.Errorf("failed to copy existing dir to temporal dir: %w", err)
			}

			vendorLocal := filepath.Join(dir, "vendor")

			if err := os.Rename(tmpVendor, vendorLocal); err != nil {
				return xerrors.Errorf("failed to symlink to vendor: %w", err)
			}

			buf := bytes.NewBuffer(nil)

			if err := command.Execute(buf, map[string]interface{}{
				"FunctionName": "pubsub-handler-" + topic,
				"Region":       region,
				"Runtime":      runtime,
				"EntryPoint":   "PubSubHandler",
				"Source":       dir,
				"Topic":        topic,
				"Options":      options,
			}); err != nil {
				return xerrors.Errorf("failed to generate gcloud functions deploy command: %w", err)
			}

			if dryRun {
				log.Printf("skipping for %s: %s", topic, buf.String())
			} else {
				log.Printf("Deploying %s started", topic)
				cmd := exec.Command("sh", "-c", buf.String())
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

				if err := cmd.Run(); err != nil {
					return xerrors.Errorf("%s failed: %w", buf.String(), err)
				}

				log.Printf("Deploying %s completed", topic)
			}

			if err := os.Rename(vendorLocal, tmpVendor); err != nil {
				return xerrors.Errorf("failed to recover vendor dir: %w", err)
			}

			return nil
		}(); err != nil {
			log.Fatalf("failed to deploy cloud functions for %s in %s: %+v", topic, funcDir, err)
		}
	}

	return nil
}
