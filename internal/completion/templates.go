package completion

// bashTemplate is the completion script template for Bash shells.
const bashTemplate = `# Bash completion for {{.ProgramName}}

_{{.ProgramName}}_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # All available flags
    opts="{{range .Flags}}--{{.Name}} {{end}}"

    # Value completions for specific flags
    case "${prev}" in
{{- range .Flags}}
{{- if and .HasValue .Values}}
        --{{.Name}})
            COMPREPLY=( $(compgen -W "{{range .Values}}{{.}} {{end}}" -- ${cur}) )
            return 0
            ;;
{{- else if and .HasValue (eq .ValueHint "DIR")}}
        --{{.Name}})
            COMPREPLY=( $(compgen -d -- ${cur}) )
            return 0
            ;;
{{- else if and .HasValue (or (eq .ValueHint "FILE") (eq .ValueHint "PATH"))}}
        --{{.Name}})
            COMPREPLY=( $(compgen -f -- ${cur}) )
            return 0
            ;;
{{- end}}
{{- end}}
    esac

    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
}

complete -F _{{.ProgramName}}_completion {{.ProgramName}}
`

// zshTemplate is the completion script template for Zsh shells.
const zshTemplate = `#compdef {{.ProgramName}}

# Zsh completion for {{.ProgramName}}

_{{.ProgramName}}() {
    local -a args
    args=(
{{- range .Flags}}
{{- if not .HasValue}}
        '--{{.Name}}[{{.Description}}]'
{{- else if .Values}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:({{range .Values}}{{.}} {{end}})'
{{- else if eq .ValueHint "DIR"}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files -/'
{{- else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files'
{{- else}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:'
{{- end}}
{{- end}}
    )
    _arguments -s -S $args
}

_{{.ProgramName}} "$@"
`

// fishTemplate is the completion script template for Fish shells.
const fishTemplate = `# Fish completion for {{.ProgramName}}

{{range .Flags -}}
{{- if not .HasValue}}
# {{.Description}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}'
{{else if .Values}}
# {{.Description}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -x -a '{{range .Values}}{{.}} {{end}}'
{{else if eq .ValueHint "DIR"}}
# {{.Description}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r -a '(__fish_complete_directories)'
{{else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
# {{.Description}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r -F
{{else}}
# {{.Description}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r
{{end}}
{{end -}}
`
