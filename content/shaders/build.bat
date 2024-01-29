for /f %%a IN ('dir /b') do call glslc %%a -o spv/%%a.spv
for /f %%a IN ('dir /b') do call glslc %%a -o spv/%%a.oit.spv -DOIT

pause