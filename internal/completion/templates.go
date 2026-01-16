package completion

// bashTemplate is the completion script template for Bash shells.
const bashTemplate = `# Bash completion for {{.ProgramName}}

_{{.ProgramName}}_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Subcommands
    local subcommands="{{range .Subcommands}}{{.Name}} {{end}}"

    # Global flags
    local global_opts="{{range .Flags}}--{{.Name}} {{end}}"

    # If we're at position 1 and not after a flag, complete subcommands or global flags
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${subcommands} ${global_opts}" -- ${cur}) )
        return 0
    fi

    # Detect which subcommand we're in
    local subcommand=""
    for word in "${COMP_WORDS[@]:1:COMP_CWORD-1}"; do
        if [[ " ${subcommands} " =~ " ${word} " ]]; then
            subcommand="${word}"
            break
        fi
    done

    # Value completions for global flags
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

    # Subcommand-specific completions
    case "${subcommand}" in
{{- range .Subcommands}}
        {{.Name}})
            local subcmd_opts="{{range .Flags}}--{{.Name}} {{end}}"

            # Value completions for subcommand flags
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

            COMPREPLY=( $(compgen -W "${subcmd_opts}" -- ${cur}) )
            return 0
            ;;
{{- end}}
    esac

    # Default: complete global flags
    COMPREPLY=( $(compgen -W "${global_opts}" -- ${cur}) )
}

complete -F _{{.ProgramName}}_completion {{.ProgramName}}
`

// zshTemplate is the completion script template for Zsh shells.
const zshTemplate = `#compdef {{.ProgramName}}

# Zsh completion for {{.ProgramName}}

_{{.ProgramName}}() {
    local line state

    _arguments -C \
{{- range .Flags}}
{{- if not .HasValue}}
        '--{{.Name}}[{{.Description}}]' \
{{- else if .Values}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:({{range .Values}}{{.}} {{end}})' \
{{- else if eq .ValueHint "DIR"}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files -/' \
{{- else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files' \
{{- else}}
        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:' \
{{- end}}
{{- end}}
        '1: :->cmds' \
        '*::arg:->args'

    case "$state" in
        cmds)
            local -a subcommands
            subcommands=(
{{- range .Subcommands}}
                '{{.Name}}:{{.Description}}'
{{- end}}
            )
            _describe 'subcommand' subcommands
            ;;
        args)
            case $line[1] in
{{- range .Subcommands}}
                {{.Name}})
                    _arguments \
{{- range .Flags}}
{{- if not .HasValue}}
                        '--{{.Name}}[{{.Description}}]' \
{{- else if .Values}}
                        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:({{range .Values}}{{.}} {{end}})' \
{{- else if eq .ValueHint "DIR"}}
                        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files -/' \
{{- else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
                        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:_files' \
{{- else}}
                        '--{{.Name}}[{{.Description}}]:{{.ValueHint}}:' \
{{- end}}
{{- end}}
                    ;;
{{- end}}
            esac
            ;;
    esac
}

_{{.ProgramName}} "$@"
`

// fishTemplate is the completion script template for Fish shells.
const fishTemplate = `# Fish completion for {{.ProgramName}}

# Subcommands
{{range .Subcommands -}}
complete -c {{$.ProgramName}} -f -n '__fish_use_subcommand' -a '{{.Name}}' -d '{{.Description}}'
{{end}}

# Global flags
{{range .Flags -}}
{{- if not .HasValue}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}'
{{else if .Values}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -x -a '{{range .Values}}{{.}} {{end}}'
{{else if eq .ValueHint "DIR"}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r -a '(__fish_complete_directories)'
{{else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r -F
{{else}}
complete -c {{$.ProgramName}} -l {{.Name}} -d '{{.Description}}' -r
{{end}}
{{end}}

# Subcommand-specific flags
{{range $subcmd := .Subcommands -}}
{{range $subcmd.Flags -}}
{{- if not .HasValue}}
complete -c {{$.ProgramName}} -n '__fish_seen_subcommand_from {{$subcmd.Name}}' -l {{.Name}} -d '{{.Description}}'
{{else if .Values}}
complete -c {{$.ProgramName}} -n '__fish_seen_subcommand_from {{$subcmd.Name}}' -l {{.Name}} -d '{{.Description}}' -x -a '{{range .Values}}{{.}} {{end}}'
{{else if eq .ValueHint "DIR"}}
complete -c {{$.ProgramName}} -n '__fish_seen_subcommand_from {{$subcmd.Name}}' -l {{.Name}} -d '{{.Description}}' -r -a '(__fish_complete_directories)'
{{else if or (eq .ValueHint "FILE") (eq .ValueHint "PATH")}}
complete -c {{$.ProgramName}} -n '__fish_seen_subcommand_from {{$subcmd.Name}}' -l {{.Name}} -d '{{.Description}}' -r -F
{{else}}
complete -c {{$.ProgramName}} -n '__fish_seen_subcommand_from {{$subcmd.Name}}' -l {{.Name}} -d '{{.Description}}' -r
{{end}}
{{end -}}
{{end -}}
`
