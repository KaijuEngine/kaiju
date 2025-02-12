---
title: Building new fonts | Kaiju Engine
---

# Building new fonts
Kaiju uses MSDF (multi-channel signed distance field) fonts for rendering text. This allows for high quality text rendering at any size. Other forms of fonts (such as bitmap) is not supported by default, you'll need to add support for fonts like that yourself if you need to [[1](#notes)].

## Building MSDF fonts
To build new font's you'll need the `msdf-atlas-gen` tool, which can be [found here](https://github.com/Chlumsky/msdf-atlas-gen/releases). Place this executable into the `bin` folder of the Kaiju repository (you may need to create this folder). Also in this folder, create a folder for the font face you'd like to convert. For example, if you'd like to convert the OpenSans font, create a folder called `OpenSans`. Inside this folder, place the TTF files for the font. Lastly, you'll need to create a text file named `charset.txt` within your font folder. This text file should have all of the characters you need out of your font. Check out the sample `charset.txt` file in the `content/editor/fonts/charset.txt` file for an example. Make note of double quotes on the ends, the escaped characters, and the UTF-8 file format.

Once you've done this setup work, you can run the following command from within the `src` folder:

```bash
go run ./generators/msdf/main.go OpenSans
```

## Using MSDF fonts
You'll need to replace `OpenSans` with whatever your folder name is. Once this process completes, it will create a new folder within your font folder `out` which has all the `.bin` and `.png` files for your font. So this would be `OpenSans/out` in our example.

Copy these files over to the `content/fonts` folder or the `content/editor/fonts` folder to begin using them. At this point you can create a `const` wherever you need it that is `rendering.FontFace` (a `string` alias). This is what you will pass into the font/label code to bind your font face for use.

## Notes
[1] The font system uses a mapping of character->glyph so it has everything you need to support bitmap fonts. You'll need to change the shader that is used by the font system to support bitmap fonts. You'll also need to make a custom build of the `.bin` file to go along with your font, see how the `src/generators/msdf/main.go` builds this binary for more information.