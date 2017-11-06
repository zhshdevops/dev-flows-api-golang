#!/bin/sh
#
# Licensed Materials - Property of tenxcloud.com
# (C) Copyright 2016 TenxCloud. All Rights Reserved.
#
set -e

# Default code repository that contains the source code 
cd $APP_CODE_REPO


if [ -z ${DOCKERFILE_NAME} ]; then
    DOCKERFILE_NAME="Dockerfile"
fi

DockerfilePath=.${DOCKERFILE_PATH-/}${DOCKERFILE_NAME}

echo "Using Dockerfile from ${DockerfilePath}"
# Check if Dockerfile exists
if [ ! -f ${DockerfilePath} ]; then
    echo "   Error: no ${DockerfilePath} found, please add this file and try again."
    exit 1
fi

if [ ! -z "$IMAGE_NAME" ]; then
    # Login using user/password for pull/push, read from user secret
    UserMatcher=".[\"$REGISTRY\"].Username"
    PasswordMatcher=".[\"$REGISTRY\"].Password"
    USERNAME=`jq $UserMatcher /docker/secret/.dockercfg -r`
    PASSWORD=`jq $PasswordMatcher /docker/secret/.dockercfg -r`

    if [ ! -z "$USERNAME" ] && [ ! -z "$PASSWORD" ]; then
        #REGISTRY=$(echo $IMAGE_NAME | tr "/" "\n" | head -n1 | grep "\." || true)
        echo "   Logging into registry ..."
        docker login -u $USERNAME -p $PASSWORD $REGISTRY >> /silentAction.log
    else
        echo "   WARNING: no user credentials for pushing/pulling"
    fi

    echo "=> Building image $IMAGE_NAME"
    docker build --pull --rm=true --force-rm $BUILD_CACHE_OPTION -t $REGISTRY/$IMAGE_NAME -f $DockerfilePath .

    if [ "$PUSH_ON_COMPLETE" = "off" ]; then
        echo "=> Skip to push image to Registry Server"
    else
        echo "=>  Pushing image $IMAGE_NAME"
        # Push to docker registry
        docker push "$REGISTRY/$IMAGE_NAME"

        # Remove the images
        docker rmi -f "$REGISTRY/$IMAGE_NAME" >> /silentAction.log
    fi
else
    echo "   WARNING: no \$IMAGE_NAME specified - will not do the build and push action."
fi
