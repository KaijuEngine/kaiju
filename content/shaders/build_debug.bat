for /f %%a IN ('dir /b') do call glslc %%a -o spv/%%a.spv -g
for /f %%a IN ('dir /b') do call glslc %%a -o spv/%%a.spv -g -DOIT

pause