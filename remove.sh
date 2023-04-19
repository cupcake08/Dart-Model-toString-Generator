#!/bin/bash

source header.sh

read -p "Enter folder path: " dir

# Checks if folder exist
if ! [ -d $dir ]; then
    printf "${RED}Folder Doesn't Exist\n"
    exit 1
fi

# Get all generated file paths recursively
gen_file_paths=()
while IFS= read -r -d '' file; do
    gen_file_paths+=("$file")
done < <(find $dir -type f -name "*_toString.g.dart" -print0)

# Prompt the user for confirmation
echo "The following files will be deleted:"
for file in ${gen_file_paths[@]}; do
    printf "${YELLOW}> ${GREEN}$file\n${NC}"
done
echo
read -p "Are you sure you want to delete these files? [y/n] " choice

if [ $choice == "y" ]; then
    # Remove the files
    for file_path in ${gen_file_paths[@]}; do
        rm $file_path
    done
    while IFS= read -r -d '' file; do
        line=$(grep -n "part" "$file" | head -n 1 | cut -d':' -f1)
        if [ -n "$line" ]; then
            sed -i "${line}d;$(($line+1))d" $file
        else
            echo "No match found file > ${file}"
        fi
    done < <(find $dir -type f -name "*.dart" -print0)
else
    echo "Operation Aborted!!"
fi

