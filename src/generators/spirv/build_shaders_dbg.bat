cd ../../../content/renderer/src
for /f %%a IN ('dir /b /a-d *.vert,*.frag') do call "../../src/generators/spirv/spirv.exe" -i="%%a" -o="spv/" -d=true