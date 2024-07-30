#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 pull [--backup] <images.txt>"
    echo "       $0 load --file <backup.zip>"
    exit 1
}

# Function to pull images and optionally back them up
pull_images() {
    images_file=$1
    backup=$2

    # Check if the images file exists
    if [ ! -f "$images_file" ]; then
        echo "File $images_file not found!"
        exit 1
    fi

    # Pull each image in the list
    while read -r image; do
        docker pull $image
    done < "$images_file"

    if [ "$backup" == "true" ]; then
        # Define the output tar file and zip file
        tarfile="docker_images.tar"
        zipfile="docker_images.zip"

        # Create a temporary directory to store individual tar files
        mkdir -p temp_docker_images

        # Loop through each image in the list and save it
        while read -r image; do
            imagename=$(echo $image | tr '/' '_' | tr ':' '_')
            docker save $image -o temp_docker_images/${imagename}.tar
        done < "$images_file"

        # Combine all the individual tar files into one
        tar -cvf $tarfile -C temp_docker_images .

        # Remove the temporary directory
        rm -rf temp_docker_images

        # Compress the tar file into a zip file
        zip $zipfile $tarfile

        # Remove the tar file
        rm $tarfile

        echo "Docker images saved and compressed to $zipfile"
    fi
}

# Function to load images from a backup file
load_images() {
    zipfile=$1

    # Check if the zip file exists
    if [ ! -f "$zipfile" ]; then
        echo "File $zipfile not found!"
        exit 1
    fi

    # Define the tar file name
    tarfile="docker_images.tar"

    # Unzip the zip file
    unzip $zipfile

    # Untar the combined tar file into a temporary directory
    mkdir -p temp_docker_images
    tar -xvf $tarfile -C temp_docker_images

    # Loop through each tar file in the temporary directory and load the image
    for image_tar in temp_docker_images/*.tar; do
        docker load -i $image_tar
    done

    # Clean up the temporary files
    rm -rf temp_docker_images
    rm $tarfile

    echo "Docker images have been loaded into the local registry"
}

# Main script logic
if [ $# -lt 1 ]; then
    usage
fi

command=$1
shift

case "$command" in
    pull)
        if [ $# -lt 1 ]; then
            usage
        fi

        backup=false
        images_file=""

        while [ $# -gt 0 ]; do
            case "$1" in
                --backup)
                    backup=true
                    ;;
                *)
                    images_file=$1
                    ;;
            esac
            shift
        done

        if [ -z "$images_file" ]; then
            usage
        fi

        pull_images $images_file $backup
        ;;
    load)
        if [ "$1" != "--file" ] || [ -z "$2" ]; then
            usage
        fi

        zipfile=$2
        load_images $zipfile
        ;;
    *)
        usage
        ;;
esac
