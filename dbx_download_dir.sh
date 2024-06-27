#!/bin/bash

# Function to create a directory if it doesn't exist
ensure_directory() {
  local path="$1"
  if [ ! -d "$path" ]; then
    mkdir -p "$path"
  fi
}

# Function to download files and replicate directory structure
download_files() {
  local path="$1"
  local destination="$2"
  local depth="$3"
  local max_parallel_jobs="$4"
  local top_name="${path##*/}"
  local local_dest_path="$destination/$top_name"
  local tabs=""
  for ((i=0; i<depth; i++)); do
    tabs+="    "
  done

  # Check if the item is not a file (does not contain a dot in the filename)
  if [[ "$path" != *.* ]]; then
    ensure_directory "$local_dest_path"
    echo -e "${tabs}Folder: $top_name"

    # Get the total number of items in the directory without leading whitespaces
    local total_items=$(dbxcli ls -l "$path" 2>/dev/null | awk 'NR>1 {print $NF}' | wc -l | awk '{$1=$1};1')
    local current_item=1
    tabs+="    "
    # Iterate over items in the directory
    dbxcli ls -l "$path" | awk 'NR>1 {print $NF}' | while read -r item; do
      local local_name="${item##*/}"
      echo -e "${tabs}Item $current_item of $total_items"
      local new_depth=$((depth+1))
      download_files "$item" "$local_dest_path" "$new_depth" "$max_parallel_jobs"
      ((current_item++))
    done
  else
    # If it's a file (contains a dot), use dbxcli get (output silenced)
    echo -e "${tabs}Downloading: $top_name"
    dbxcli get "$path" "$local_dest_path" > /dev/null 2>&1
  fi
}

echo "Download completed."

    # Limit the number of parallel jobs
    while (( $(jobs -r | wc -l) >= max_parallel_jobs )); do
      sleep 1
    done
  fi
}

# Set default max_parallel_jobs to 8 if not provided
max_parallel_jobs="${3:-16}"

# Download files and replicate directory structure
download_files "$1" "$2" 0 "$max_parallel_jobs"

# Wait for all background processes to finish
wait

echo "Download completed."

