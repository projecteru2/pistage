package command

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v4"
)

// RenderCommand renders commandTemplate with the given arguments using Jinja
// "env" and "vars" will be injected into context and render the template,
// if they are also defined in arguments, arguments will be overridden.
func RenderCommand(commandTemplate string, arguments, env, vars map[string]string) (string, error) {
	tmpl, err := pongo2.FromString(commandTemplate)
	if err != nil {
		return "", err
	}

	context := pongo2.Context{}
	for k, v := range arguments {
		context[k] = v
	}
	context["vars"] = vars
	context["env"] = env

	// first render, replace arguments in template
	out, err := tmpl.Execute(context)
	if err != nil {
		return "", err
	}

	// second render, replace vars and env in arguments
	tmpl2, err := pongo2.FromString(out)
	if err != nil {
		return "", err
	}
	return tmpl2.Execute(context)
}

var shell = `{% for cmd in commands %}{{ cmd | safe }}
{% endfor %}`

func RenderShell(commands []string) (string, error) {
	tmpl, err := pongo2.FromString(shell)
	if err != nil {
		return "", err
	}
	return tmpl.Execute(pongo2.Context{"commands": commands})
}

func RenderEnvironmentForSSH(envs map[string]string) string {
	lines := []string{}
	for name, value := range envs {
		lines = append(lines, fmt.Sprintf("export %s=%s", name, value))
	}
	return strings.Join(lines, "\n")
}

// EmptyWorkloadCommand returns the command to execute in an empty workload.
// Currently we use a sleep timeout command to achieve this.
// Interestingly, k8s uses pause and GitHub Action uses tail -f /dev/null.
// Great minds think alike! ( ... ( '-' )ノ)`-' )
func EmptyWorkloadCommand(timeout int) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("sleep %d", timeout)}
}

func regulateEnv(env string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(env), ".", "_"), "-", "_")
}

// ToEnvironmentList transfers an environment map to a key=value list
// key will be in upper case, and . in key will be replaced by _
func ToEnvironmentList(env map[string]string) []string {
	var envs []string
	for key, value := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", regulateEnv(key), value))
	}
	return envs
}

func PreparePistageEnvs(envs map[string]string) map[string]string {
	prepared := make(map[string]string, len(envs))
	for key, value := range envs {
		prepared[fmt.Sprintf("PISTAGE_ENV_VAR_%s", key)] = value
	}
	return prepared
}

// MergeVariables merges higherPriority into lowerPriority,
// will override the values in lowerPriority.
func MergeVariables(lowerPriority, higherPriority map[string]string) map[string]string {
	m := map[string]string{}
	for k, v := range lowerPriority {
		m[k] = v
	}
	for k, v := range higherPriority {
		m[k] = v
	}
	return m
}
