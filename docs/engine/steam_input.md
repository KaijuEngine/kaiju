# Steam Input

Kaiju's Steam Input integration is action-based. It does not replace
`platform/hid.Controller`, which remains available for legacy Xbox-style
gamepad input and for builds that do not initialize Steam.

Game code that imports `platform/steam` should normally live in a file with a
`//go:build steam` constraint, with a non-Steam implementation in a matching
`//go:build !steam` file. This keeps ordinary builds independent of the
Steamworks runtime library.

Build the game with the `steam` tag to enable the Steam bootstrap:

```bash
go build -tags="steam,debug" -o ../kaijuengine.com ./
```

The bootstrap initializes `SteamAPI`, initializes `ISteamInput`, and calls
`SteamAPI_RunCallbacks` before game updates each frame. Shutting down the host
shuts down Steam Input before shutting down SteamAPI.

## Action manifest

Steam Input reads semantic actions from a VDF action manifest. A game should
normally define separate action sets for menu and gameplay controls:

```text
"Action Manifest"
{
    "actions"
    {
        "menu"
        {
            "title" "#Menu"
            "legacy_set" "0"
            "Button"
            {
                "menu_up" "#MenuUp"
                "menu_down" "#MenuDown"
                "confirm" "#Confirm"
                "cancel" "#Cancel"
            }
        }
        "gameplay"
        {
            "title" "#Gameplay"
            "legacy_set" "0"
            "StickPadGyro"
            {
                "camera_pan"
                {
                    "title" "#CameraPan"
                    "input_mode" "joystick_move"
                }
            }
            "Button"
            {
                "pause" "#Pause"
                "primary_action" "#PrimaryAction"
            }
        }
    }
    "localization"
    {
        "english"
        {
            "Menu" "Menu"
            "MenuUp" "Up"
            "MenuDown" "Down"
            "Confirm" "Confirm"
            "Cancel" "Cancel"
            "Gameplay" "Gameplay"
            "CameraPan" "Pan Camera"
            "Pause" "Pause"
            "PrimaryAction" "Primary Action"
        }
    }
}
```

During local development, install `game_actions_<AppID>.vdf` in Steam's
`controller_config` directory or select a bundled manifest after Steam
initializes:

```go
if steam.SteamInput.IsInitialized() {
    if !steam.SteamInput.SetActionManifestFilePath("content/steam_input_manifest.vdf") {
        slog.Warn("failed to select Steam Input manifest")
    }
}
```

Production builds should include the action manifest and official controller
configurations in the depot. Select **Custom Configuration (Bundled with
Game)** on the app's Steam Input partner settings, set the manifest's relative
path, opt in the supported controller types, and publish the configuration.

## Resolving and polling actions

Resolve handles once after Steam initializes. Resolving a digital or analog
action also registers it for automatic frame polling:

```go
type gameInput struct {
    menuSet   steam.InputActionSetHandle
    gameplay  steam.InputActionSetHandle
    confirm   steam.InputDigitalActionHandle
    cancel    steam.InputDigitalActionHandle
    cameraPan steam.InputAnalogActionHandle
}

func newGameInput() gameInput {
    return gameInput{
        menuSet:   steam.SteamInput.ActionSetHandle("menu"),
        gameplay:  steam.SteamInput.ActionSetHandle("gameplay"),
        confirm:   steam.SteamInput.DigitalActionHandle("confirm"),
        cancel:    steam.SteamInput.DigitalActionHandle("cancel"),
        cameraPan: steam.SteamInput.AnalogActionHandle("camera_pan"),
    }
}
```

Activate the action set that matches the current game state. It is safe to
activate a set repeatedly:

```go
for _, controller := range steam.SteamInput.Controllers() {
    steam.SteamInput.ActivateActionSet(controller, input.menuSet)

    confirm := steam.SteamInput.DigitalAction(controller, input.confirm)
    if confirm.Pressed {
        acceptSelection()
    }
}
```

When gameplay starts, activate its set before reading gameplay actions:

```go
for _, controller := range steam.SteamInput.Controllers() {
    steam.SteamInput.ActivateActionSet(controller, input.gameplay)
    pan := steam.SteamInput.AnalogAction(controller, input.cameraPan)
    if pan.Active {
        moveCamera(pan.X, pan.Y)
    }
}
```

`DigitalActionData` exposes the native `State` and `Active` values plus
engine-generated `Pressed`, `Held`, and `Released` transitions. Registered
actions are polled after Steam callbacks and before game update functions.

`SteamInput.RunFrame` is available for applications that need an explicit,
lower-latency refresh immediately before reading input. Normal game code should
use the automatic callback-driven refresh to avoid polling twice.

## Prompts and device features

Use action origins instead of assuming Xbox button labels:

```go
origins := steam.SteamInput.DigitalActionOrigins(controller, input.menuSet, input.confirm)
if len(origins) > 0 {
    label := steam.SteamInput.OriginName(origins[0])
    glyphPath := steam.SteamInput.GlyphSVG(
        origins[0],
        steam.InputGlyphStyleKnockout|steam.InputGlyphStyleSolidABXY,
    )
    _, _ = label, glyphPath
}
```

The system also exposes:

- controller model detection and legacy gamepad-index correlation;
- Steam's controller binding panel;
- active action-set layers;
- motion data using Kaiju `matrix` types;
- vibration, trigger vibration, and Steam Deck/Steam Controller haptics.

When native Steam Input and legacy joystick input are both enabled, avoid
processing the same controller through both paths. Use
`GamepadIndexForController` to correlate Steam Input controllers with legacy
gamepad slots, or configure selected controller types to use only one input
path.

Valve's setup and publishing workflow is documented at
<https://partner.steamgames.com/doc/features/steam_controller/getting_started_for_devs>.
