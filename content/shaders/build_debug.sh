shopt -s nullglob
for i in *.vert *.frag *.geom *.tesc *.tese; do
	echo "Compiling $i (debug)"
	glslc $i -o spv/$i.spv -g
	glslc $i -o spv/$i.spv -g -DOIT
done
