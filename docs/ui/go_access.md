---
title: Kaiju Engine | Go Access
---

# Accessing the UI elements from Go
At some point you'll want to access the UI that you've designed in HTML from Go. To do this you'll want to take that returned document from `markup.DocumentFromHTMLAsset` and access some helper functions or the list of elements.

## GetElementById
In our [Writing](writing.md) example, we created a `div` with an id `nameList`. This allows us to get access to all of the elements inside of that `div` by calling `GetElementById` on the document.

```go
//...
doc, err := markup.DocumentFromHTMLAsset(host, "ui/tests/binding.html", data, nil)
list, ok := doc.GetElementById("nameList")
```

In this case `ok` will be `false` if the element could not be found. Otherwise `list` will be the `DocElement` for that panel. You can access the child entities from this panel and go through all the child contents that way.

## Other accessors
You can access other elements by class, tag, or group. Class and tag are the classic ways to access elements in HTML, and group is a way to access a group of elements that have the same value in the `group` html attribute.

For example, the divs with ids `one`, `two`, and `three` all have the same group value of `group1`. You can access all of these divs by calling `GetElementsByGroup` on the document.

```html
<!-- ... -->
<div>
	<div id="one" class="red" group="group1"></div>
</div>
<div id="two" class="green" group="group1"></div>
<div id="three" class="blue" group="group1"></div>
<!-- ... -->
```