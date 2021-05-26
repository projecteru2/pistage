package variable

import (
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v4"
)

var (
	varsRe = regexp.MustCompile(`(?U){{\s*(\$env|\$vars).*}}`)

	phistageEnvVarName  = "__phistage_env__"
	phistageVarsVarName = "__phistage_vars__"
)

// ReplaceVariables replaces variables startswith $ provided by phistage
// to phistage private variables, which will later be renderred.
func ReplaceVariables(t string) string {
	return varsRe.ReplaceAllStringFunc(t, func(m string) string {
		r := m
		r = strings.Replace(r, "$env", phistageEnvVarName, 1)
		r = strings.Replace(r, "$vars", phistageVarsVarName, 1)
		return r
	})
}

// BuildTemplateContext uses arguments, env, and vars to build pongo2 context
// for rendering the template
func BuildTemplateContext(arguments, envs, vars map[string]string) pongo2.Context {
	context := pongo2.Context{
		phistageEnvVarName:  envs,
		phistageVarsVarName: vars,
	}
	for k, v := range arguments {
		context[k] = v
	}
	return context
}

// RenderArguments renders arguments with envs and vars.
// We support variables in arguments, so before we send the arguments to executable,
// we need to first render arguments with variables.
func RenderArguments(arguments, envs, vars map[string]string) (map[string]string, error) {
	context := pongo2.Context{"env": envs, "vars": vars}
	r := map[string]string{}

	for k, v := range arguments {
		t, err := pongo2.FromString(v)
		if err != nil {
			return nil, err
		}

		o, err := t.Execute(context)
		if err != nil {
			return nil, err
		}
		r[k] = o
	}
	return r, nil
}
