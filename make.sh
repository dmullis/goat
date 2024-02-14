#! /bin/sh
#
# Run all tests, and all pre-compilation build steps.
# Certain output files should be committed to the SCM archive.
#
# Recall that an end-user eventually installing with 'go get ...'
# will trigger a compilation from source within the local environment.
#  XX  Give this file a more descriptive name.

set -e
set -x
usage () {
    printf "%s\n\n" "$*"
    printf "usage: %s [-g GitHub_Username] [-w]\n" ${0##*/}
    printf "\t%s\t%s\n" ""
    printf "\t%s\t%s\n" "$*"
    exit 1
}

# Define colors for SVG ~foreground~ seen on Github front page.
svg_color_dark_scheme="#EEF"
svg_color_light_scheme="#011"
github_blue_color="#2F81F7"

# GOMOD=$(go env GOMOD)
# from_username=${GOMOD##*github.com/}
# githubuser=${from_username%%/*}
#
# X  Is it acceptable to push to a PR branch files that refer to the owner's main branch?
githubuser=blampe

TEST_ARGS=

while getopts hg:iw flag
do
    case $flag in
        h)  usage "";;
        g)  githubuser=${OPTARG};;  # Override guess based on GOMOD
	w)  TEST_ARGS=${TEST_ARGS}" -regenerate";;
        \?) usage "unrecognized option flag";;
    esac
done

tmpl_expand_GH () {
    go run ./cmd/tmpl-expand Root="." GithubUser=${githubuser} "$@"
}

# SVG examples/ regeneration.
#
# If the command fails due to expected changes in SVG output, rerun
# this script with "TEST_ARGS=-regenerate" first on the command line.
# X  Results are used as "golden" standard for GitHub-side regression tests --
#    so arguments here must not conflict with those in "test.yml".
#   XX  How to share a single arg list shared between the two i.e. "DRY"?
go test -run . -v \
   ${TEST_ARGS}

# Dump all optional args available for `go test`:
#    go test -run . -v -args -h
# Run test binary under debugger:
#    dlv test -- -test.v -regenerate

# Illustrate a workaround for lack of support in certain browsers e.g. Safari for
# inheritance of CSS property 'color-scheme' from <img> elements downward to nested
# <svg> elements.
#  - https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme
#
# XX  Build an executable locally, for ease of debugging,
# thereby preserving the pre-hoc binary in some GOPATH dir for ease if comparison.
go run ./cmd/goat <examples/trees.txt \
   -svg-color-dark-scheme ${github_blue_color} \
   -svg-color-light-scheme ${github_blue_color} \
   >trees.mid-blue.svg

# build README.md
# XX  Complication: filenames want to use '-', but template identifiers disallow them.
#      => $ename .txt files to use '_'?
# This uses the 'ValueFilePath' functionality of cmd/tmpl-expand
filename_pairs=$(grep --only-matching -E '[_a-z]+_txt' README.md.tmpl |
		     sed 's:\(.*\)_txt: ./examples/\1.txt ./examples/\1.svg:' |
		     tr '_' '-')
tmpl_expand_GH <README.md.tmpl >README.md $filename_pairs

# '-d' writes ./awkvars.out
cat *.go |
    awk '
        /[<]goat[>]/ {p = 1; next}
        /[<][/]goat[>]/ {p = 0; next}
        p > 0 {print}' |
    tee goat.txt |
    go run ./cmd/goat \
	-svg-color-dark-scheme ${svg_color_dark_scheme} \
	-svg-color-light-scheme ${svg_color_light_scheme} \
	>goat.svg

# Render to HTML, for local inspection.
./markdown_to_html.sh README.md >README.html
./markdown_to_html.sh CHANGELOG.md >CHANGELOG.html

printf "\nTo install in local GOPATH:\n\t%s\n" "go install ./cmd/goat"
