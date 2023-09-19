#!/usr/bin/env bash
set -euxo pipefail

HELP='Usage: ./package_ev2.sh ${NAME}

Will copy the build output (binary + dockerfile) and ev2 configuration,
for image push to the drop folder.
'

if [ "$#" -ne 1 ]; then
    echo $HELP
    exit 1
fi

BASH_ROOT="$(dirname "${BASH_SOURCE[0]}")/.."

cd "$BASH_ROOT"

EV2_OUTPUT="/source/out/${1}/ServiceGroupRoot"
IMG_NAME=${1}

mkdir -p "$EV2_OUTPUT/bin"

apt update && apt install -y gettext-base git rsync

rsync -av "/source/scripts/ev2/image-build/." "$EV2_OUTPUT" --exclude "push-image.sh"
mkdir -p "/source/out/bin/dist/${1}_linux_amd64/${1}"

chmod +x "/source/dist/${1}_linux_amd64/${1}"
cp -a "/source/images/Dockerfile" "/source/out/bin/Dockerfile"
cp -a "/source/dist/${1}_linux_amd64/${1}" "/source/out/bin/dist/${1}_linux_amd64/${1}"
cp -a "/source/scripts/ev2/image-build/push-image.sh" "/source/out/bin/push-image.sh"

tar -cf "$EV2_OUTPUT/bin/acrpush.tar" -C "/source/out/bin" .
rm -rf "/source/out/bin"

build_date=$(date +'%Y%m%d')
build_time=$(date +'%-H%M%S' | sed 's/^0*//')
git_branch="$BUILD_SOURCEBRANCHNAME"
git_sha="$(git rev-parse --short=8 $BUILD_SOURCEVERSION 2>/dev/null)"
BUILD_ID="${git_branch:-newrepo}.${build_date}-${git_sha:-00000000}"
echo "IMAGE_NAME=\"${IMG_NAME}\" BUILD_ID=\"${BUILD_ID}\""
IMAGE_NAME="${IMG_NAME}" BUILD_ID="${BUILD_ID}" envsubst < "$EV2_OUTPUT/ScopeBindings.json" > tmp.json
mv tmp.json "$EV2_OUTPUT/ScopeBindings.json"
