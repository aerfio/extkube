package main

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"golang.design/x/clipboard"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	kubectlconfig "k8s.io/kubectl/pkg/cmd/config"
)

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	if err := clipboard.Init(); err != nil {
		return err
	}

	context := ""
	help := false
	list := false
	pflag.StringVarP(&context, "context", "c", "", "kube context to extract")
	pflag.BoolVarP(&help, "help", "h", false, "prints help message")
	pflag.BoolVarP(&list, "list", "l", false, "list available k8s contexts. Equivalent to 'kubectl config get-contexts'")
	pflag.Parse()

	if help {
		pflag.Usage()
		return nil
	}

	if list {
		cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().GetStartingConfig()
		if err != nil {
			return err
		}

		fmt.Println(strings.Join(slices.Sorted(maps.Keys(cfg.Contexts)), "\n"))
		return nil
	}

	if context == "" {
		return fmt.Errorf("please provide --context/-c flag")
	}

	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	opts := &kubectlconfig.ViewOptions{
		PrintFlags:   genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme).WithDefaultOutput("yaml"),
		ConfigAccess: clientcmd.NewDefaultClientConfigLoadingRules(),
		Flatten:      true,
		Minify:       true,
		Merge:        cliflag.True,
		RawByteData:  true,
		Context:      context,
		IOStreams: genericclioptions.IOStreams{
			Out:    buf,
			ErrOut: errBuf,
		},
	}
	printer, err := opts.PrintFlags.ToPrinter()
	if err != nil {
		return fmt.Errorf("failed to create printer from kubectl config: %s", err)
	}
	opts.PrintObject = printer.PrintObj

	if err := opts.Run(); err != nil {
		return fmt.Errorf("failed to run the equivalent of `kubectl config view` command: %s", err)
	}
	if cmdStdErr := errBuf.String(); cmdStdErr != "" {
		return fmt.Errorf("stderr of this command was not empty: %s", cmdStdErr)
	}

	file, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %s", err)
	}
	defer file.Close()

	if _, err := file.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to file %q: %s", file.Name(), err)
	}

	msg := fmt.Sprintf("export KUBECONFIG=%q", file.Name())
	clipboard.Write(clipboard.FmtText, []byte(msg))

	fmt.Println("Copied:")
	fmt.Println(color.HiGreenString(msg))
	fmt.Println("to clipboard")

	return nil
}
