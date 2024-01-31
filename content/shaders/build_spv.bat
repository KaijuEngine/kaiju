@ECHO OFF

IF "%1"=="-d" (
	SET DEBUG=1
	ECHO Compiling in debug mode
) ELSE (
	SET DEBUG=0
	ECHO Compiling in release mode
)

FOR /f %%A IN ('dir /b *.vert') DO (
	ECHO Compiling %%A
	IF "%DEBUG%"=="1" (
		call glslc %%A -o spv/%%A.spv -g
	) ELSE (
		call glslc %%A -o spv/%%A.spv
	)
)

FOR /f %%A IN ('dir /b *.frag') DO (
	ECHO Compiling %%A
	IF "%DEBUG%"=="1" (
		call glslc %%A -o spv/%%A.spv -g
	) ELSE (
		call glslc %%A -o spv/%%A.spv
	)
	FINDSTR "^\<#ifdef OIT\>" %%A > NUL 2>&1 && (
		ECHO OIT block found in %%A, compiling OIT version
		IF "%DEBUG%"=="1" (
			call glslc %%A -o spv/%%A.oit.spv -g -DOIT
		) ELSE (
			call glslc %%A -o spv/%%A.oit.spv -DOIT
		)
	)
)
