#! /bin/bash

# Previews a markdown file every time we change it. Receives a file name.
#
# ## Usage
#
#   # Make the script executable
#   chmod +x autopreview.sh
#
#   # Watch the changes of the provided file and show the preview.
#   ./autopreview.sh README.md
#
FHASH=`md5sum $1`

# Calculate the checksum of the file every five seconds.
while true; do
  NHASH=`md5sum $1`

  # If the result is different from the previous one, the content of the file
  # was changed, triggering the execution of the mdp tool to preview it.
  if [ "$NHASH" != "$FHASH" ]; then
    ./mdp -file $1
    FHASH=$NHASH
  fi

  sleep 5
done
