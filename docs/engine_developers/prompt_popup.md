---
title: Prompt popups | Kaiju Engine
---

# Prompt popups
There are times in the engine code where you need to get some sort of feedback immediately from the developer before you continue to do any actions. Usually this is in the form of an alert box with ok and cancel buttons or a input box for some text. For this, the alert package is used, we have two functions, [New](/api/editor/alert/#new) and [NewInput](/api/editor/alert/#newinput). Both of these return a channel so that you can block until the pop up is complete. below are a couple of examples of how to use these functions.

## New
```go
ok := <-alert.New("Save Changes", "You are changing stages, any unsaved changes will be lost. Are you sure you wish to continue?", "Yes", "No", host)
// ok will be true if the "Yes" (ok) button was clicked
```

## NewInput
```go
name := <-alert.NewInput("Stage Name", "Name of stage...", "", "Save", "Cancel", host)
// The result will be "" if cancel was clicked, otherwise it's the input text
```
