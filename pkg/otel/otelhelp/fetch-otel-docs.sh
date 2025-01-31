#!/usr/bin/env bash

general_file=$1
otlp_exporter_file=$2

general_path="content/en/docs/languages/sdk-configuration/general.md"
otlp_exporter_path="content/en/docs/languages/sdk-configuration/otlp-exporter.md"

otel_docs_ref="main"
otel_docs_base="https://raw.githubusercontent.com/open-telemetry/opentelemetry.io"

function fill_relative_links() {
	file=$1
	fmtfile="${file}.tmp"
	[ -f "$fmtfile" ] && rm "$fmtfile" # remove old file
	frontmatter=false
	template=false
	lastline=""
	while IFS= read -r line; do
		# Skip frontmatter
		if [[ "${line}" == "---" ]]; then
			if $frontmatter; then
				frontmatter=false
				continue
			fi
			frontmatter=true
			continue
		elif $frontmatter; then
			# Use frontmatter title as markdown title
			if [[ "${line}" == "title: "* ]]; then
				line="# ${line#"title: "}"
			else
				continue
			fi
		fi
		# Skip hugo templating
		if [[ "${line}" == "{{% "* ]]; then
			if $template; then
				template=false
				continue
			fi
			template=true
			continue
		elif $template; then
			continue
		fi

		# Skip double newlines
		if [[ -z "$lastline" ]] && [[ -z "$line" ]]; then
			continue
		fi

		# Skip otlp link definition
		if [[ "${line}" == "[otlp]: "* ]]; then
			continue
		fi

		fmtline="$line"

		# Change relative markdown links to URLs
		fmtline="${fmtline//"](/docs/"/"](https://opentelemetry.io/docs/"}"
		fmtline="${fmtline//"]: /docs/"/"]: https://opentelemetry.io/docs/"}"

		# Replace OTLP link shortcut
		fmtline="${fmtline//"[OTLP][]"/"[OTLP](https://opentelemetry.io/docs/specs/otlp/)"}"

		lastline="$fmtline"

		printf '%s\n' "$fmtline" >>"$fmtfile"
	done <"$file"
	mv "$fmtfile" "$file"
}

curl -o "$general_file" "${otel_docs_base}/refs/heads/${otel_docs_ref}/${general_path}"
curl -o "$otlp_exporter_file" "${otel_docs_base}/refs/heads/${otel_docs_ref}/${otlp_exporter_path}"

fill_relative_links "$general_file"
fill_relative_links "$otlp_exporter_file"

markdownlint-cli2 "$general_file" --fix
markdownlint-cli2 "$otlp_exporter_file" --fix
