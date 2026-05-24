# Editor Lua Scripting

Editor Lua scripts are project-trusted automation macros. They live in
`database/scripts/editor` inside the open project and are run manually from the
Scripts workspace. Each run uses a fresh Lua VM.

Scripts can use the global `editor` object directly or define `main(editor)`.
The exposed API is a curated editor automation facade plus registered math
types such as `Vec3`.

```lua
function main(editor)
	local selected = editor:Stage():Selection()
	for i = 1, #selected do
		local pos = selected[i]:Position()
		selected[i]:SetPosition(Vec3.New(pos:X() + 1, pos:Y(), pos:Z()))
	end
	editor:Log("Moved " .. #selected .. " entities")
end
```

The Lua sandbox removes direct `os`, `io`, `package`, `load`, `loadstring`, and
`dofile` access. Use `editor:Project():ReadText(path)` and
`editor:Project():WriteText(path, text)` for project-rooted file access.
