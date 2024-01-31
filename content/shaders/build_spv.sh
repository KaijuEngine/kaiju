#!/bin/bash

if [ "$1" == "-d" ]; then
	DEBUG=1
	echo "Compiling in debug mode"
else
	DEBUG=0
	echo "Compiling in release mode"
fi

for file in *.vert; do
	echo "Compiling $file"
	if [ "$DEBUG" == "1" ]; then
		glslc $file -o spv/${file}.spv -g
	else
		glslc $file -o spv/${file}.spv
	fi
done

for file in *.frag; do
	echo "Compiling $file"
	if [ "$DEBUG" == "1" ]; then
		glslc $file -o spv/${file}.spv -g
	else
		glslc $file -o spv/${file}.spv
	fi
	if grep -q "^#ifdef OIT" $file; then
		echo "OIT block found in $file, compiling OIT version"
		if [ "$DEBUG" == "1" ]; then
			glslc $file -o spv/${file}.oit.spv -g -DOIT
		else
			glslc $file -o spv/${file}.oit.spv -DOIT
		fi
	fi
done
