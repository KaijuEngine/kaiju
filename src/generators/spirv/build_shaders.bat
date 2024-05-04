cd ../../../content/shaders/
for /f %%a IN ('dir /b /a-d *.vert,*.frag') do call "../../src/generators/spirv/spirv.exe" -i="%%a" -o="spv/"