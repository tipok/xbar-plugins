#!/usr/bin/env bash

export PATH='/usr/local/bin:/usr/bin:$PATH' # ensure brew paths are in PATH
CONFIG_FILE=$HOME/.config/github-prs/config.env

if [ -f $CONFIG_FILE ]; then
    set -a
    source $CONFIG_FILE
    set +a
fi

ORGANIZATIONS=${ORGANIZATIONS:-($USER)}
GITHUB_HOST=${GITHUB_HOST:-github.com}
USERNAME=${USERNAME:-$USER}

repos=""
for org in $ORGANIZATIONS; do
    repos_org=$(gh repo list $org | awk '{split($0,a); print a[1]}')
    repos="$repos $repos_org"
done

all_prs=0
HEADER="Open PRs"
BODY=""
for d in $repos ; do
    prs=$(gh pr list --repo "$GITHUB_HOST/$d" -S 'is:pr is:open archived:false review-requested:@me' | awk -v github_host=$GITHUB_HOST -v repo=$d '{split($0,a, "\t"); print "#"a[1]" ("a[4]") "a[2]" | href=\"https://"github_host"/"repo"/pull/"a[1]"\"" }')
    number_of_psr=$(echo $prs | sed '/^\s*$/d' | wc -l | sed 's/^[[:space:]]*//')
    if [[ $number_of_psr -ne 0 ]]; then
        all_prs=$(($all_prs + $number_of_psr))
        BODY=$BODY$'\n'"$d ($number_of_psr)"
        BODY=$BODY$'\n'$prs
        BODY=$BODY$'\n---'
    fi
done

HEADER="$HEADER ($all_prs)"
if [[ $all_prs -eq 0 ]]; then
    HEADER=":white_check_mark: $HEADER"
else
    HEADER=":warning: $HEADER"
fi
echo "$HEADER"
echo "---"
echo "$BODY"
