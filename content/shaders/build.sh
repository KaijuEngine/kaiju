shopt -s nullglob
for i in *.vert *.frag *.geom *.tesc *.tese; do
	echo "Compiling $i (release)"
	glslc $i -o spv/$i.spv
	glslc $i -o spv/$i.spv -DOIT
done
