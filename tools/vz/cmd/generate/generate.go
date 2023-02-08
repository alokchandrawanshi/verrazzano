package generate

import (
	"embed"
	"fmt"
	"github.com/spf13/cobra"
	cmdhelpers "github.com/verrazzano/verrazzano/tools/vz/cmd/helpers"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/helpers"
	"os"
	"path"
	"strings"
)

const (
	CommandName = "generate"
	helpShort   = "Generate Verrazzano manifests"
	helpLong    = `Generate Verrazzano manifests for creating cluster resources`
	helpExample = `
# Generate manifests for an OCNE cluster on OCI
vz generate --template <template-name> -o vz.yaml

# Show available templates
vz generate --list
`
)

//go:embed templates
var templates embed.FS

const (
	templatesDir  = "templates"
	yamlExtension = ".yaml"

	listArg     = "list"
	templateArg = "template"
	outputArg   = "output"
)

func NewCmdGenerate(helper helpers.VZHelper) *cobra.Command {
	cmd := cmdhelpers.NewCommand(helper, CommandName, helpShort, helpLong)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCmdGenerate(cmd, args, helper)
	}

	cmd.Example = helpExample
	cmd.PersistentFlags().Bool(listArg, false, "show available templates")
	cmd.PersistentFlags().String(templateArg, "", "template to generate")
	cmd.PersistentFlags().StringP(outputArg, "o", "", "output file for generate template")
	return cmd
}

func runCmdGenerate(cmd *cobra.Command, args []string, helper helpers.VZHelper) error {
	list, err := cmd.PersistentFlags().GetBool(listArg)
	if err != nil {
		return fmt.Errorf("an error occurred while reading value for the flag %s: %s", listArg, err.Error())
	}
	if list {
		return listTemplates(helper)
	}

	t, err := cmd.PersistentFlags().GetString(templateArg)
	if err != nil {
		return fmt.Errorf("an error occurred while reading value for the flag %s: %s", templateArg, err.Error())
	}
	o, err := cmd.PersistentFlags().GetString(outputArg)
	if err != nil {
		return fmt.Errorf("an error occurred while reading value for the flag %s: %s", outputArg, err.Error())
	}
	return generateTemplate(helper, t, o)
}

func generateTemplate(helper helpers.VZHelper, t string, o string) error {
	filePath := path.Join(templatesDir, fmt.Sprintf("%s%s", t, yamlExtension))
	b, err := templates.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(helper.GetOutputStream(), "Failed to read template: %s", t)
		return listTemplates(helper)
	}

	if len(o) < 1 {
		fmt.Fprintf(helper.GetOutputStream(), string(b))
		return nil
	}
	if err := os.WriteFile(o, b, 600); err != nil {
		fmt.Fprintf(helper.GetOutputStream(), "Failed to generate template: %s", t)
		return err
	}
	fmt.Fprintf(helper.GetOutputStream(), "Generated template %s as %s\n", t, o)
	return nil
}

func listTemplates(helper helpers.VZHelper) error {
	entries, err := templates.ReadDir(templatesDir)
	if err != nil {
		return err
	}

	fmt.Fprintln(helper.GetOutputStream(), "Verrazzano Templates:")
	for _, entry := range entries {
		fmt.Fprintln(helper.GetOutputStream(), strings.Replace(entry.Name(), yamlExtension, "", -1))
	}
	return nil
}
