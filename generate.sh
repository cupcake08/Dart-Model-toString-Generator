#!/bin/bash

source header.sh

read -p "Enter folder path: " dir

# Checks if folder exist
if ! [ -d $dir ]; then
  printf "${RED}Folder Doesn't Exist\n"
  exit 1
fi

# Remove all generated files
generated_files=()
while IFS= read -r -d '' file; do
  generated_files+=("$file")
done < <(find $dir -type f -name "*.g.dart" -print0)

for gf in ${generated_files[@]}; do
  rm $gf
done

# Get all file paths recursively
file_paths=()
while IFS= read -r -d '' file; do
  file_paths+=("$file")
done < <(find $dir -type f -name "*.dart" -print0)

# Generate
./generator ${file_paths[@]}

exit 0
