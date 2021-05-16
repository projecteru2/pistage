package command

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v4"
	"github.com/projecteru2/phistage/helpers/variable"
)

// RenderCommand renders commandTemplate with the given arguments using Jinja
func RenderCommand(commandTemplate string, arguments, envs, vars map[string]string) (string, error) {
	ctmpl := variable.ReplaceVariables(commandTemplate)
	tmpl, err := pongo2.FromString(ctmpl)
	if err != nil {
		return "", err
	}

	out, err := tmpl.Execute(variable.BuildTemplateContext(arguments, envs, vars))
	if err != nil {
		return "", err
	}

	return out, nil
}

// EmptyWorkloadCommand returns the command to execute in an empty workload.
// Currently we use a sleep timeout command to achieve this.
// Interestingly, k8s uses pause and GitHub Action uses tail -f /dev/null.
// Great minds think alike! ( ... ( '-' )ノ)`-' )
func EmptyWorkloadCommand(timeout int) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("sleep %d", timeout)}
}

// ToEnvironmentList transfers an environment map to a key=value list
// key will be in upper case, and . in key will be replaced by _
func ToEnvironmentList(env map[string]string) []string {
	var envs []string
	for key, value := range env {
		key = strings.ReplaceAll(key, ".", "_")
		envs = append(envs, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}
	return envs
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
