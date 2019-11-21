#!/bin/bash

cnt=0
while IFS='' read -r line || [[ -n "$line" ]]; do
  cnt=$((cnt+1))
  length=${#line}
  if [ $cnt -eq 1 ]; then
    # Checking if subject exceeds 50 characters
    # replace case-insensitive "(bsc#)" (or []) and surrounding spaces
    # with a single space, then prune leading/trailing spaces
    line=`echo $line | sed -E 's/[([]bsc#[0-9]+[])]// '| sed -E 's/\s*$//'`
    length=${#line}
    if [ $length -gt 50 ]; then
      echo "Your subject line exceeds 50 characters (excluding the bsc# reference)."
      exit 1
    fi
    i=$(($length-1))
    last_char=${line:$i:1}
    # Last character must not have a punctuation
    if [[ ! $last_char =~ [0-9a-zA-Z] ]]; then
      echo "Last character $last_char of the subject line must not have punctuation."
      exit 1
    fi
  elif [ $cnt -eq 2 ]; then
    # Subject must be followed by a blank line
    if [ $length -ne 0 ]; then
      echo "Your subject line follows a non-empty line. Subject lines should always be followed by a blank line."
      exit 1
    fi
  else
    # Any line in body must not exceed 72 characters
    if [ $length -gt 72 ]; then
      echo "The line \"$line\" exceeds 72 characters."
      exit 1
    fi
  fi
done < "$1"
