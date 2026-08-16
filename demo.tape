Output guh.gif

Require ffmpeg
Require ttyd

Set FontFamily "Menlo"
Set FontSize 16
Set Width 1600
Set Height 900
Set Padding 16
Set Framerate 50
Set CursorBlink false
Set TypingSpeed 80ms
Set Theme { "name": "guh-dark", "background": "#121212", "foreground": "#eeeeee", "black": "#121212", "red": "#d75f5f", "green": "#87d787", "yellow": "#ffd75f", "blue": "#87afff", "magenta": "#af87ff", "cyan": "#5fd7ff", "white": "#eeeeee", "brightBlack": "#585858", "brightRed": "#ff5f5f", "brightGreen": "#87ff87", "brightYellow": "#ffd700", "brightBlue": "#87afff", "brightMagenta": "#d787ff", "brightCyan": "#5fd7ff", "brightWhite": "#ffffff", "cursor": "#eeeeee", "selection": "#3a3a3a" }

Env TERM "xterm-256color"
Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"

# Launch off-camera so the GIF is just the TUI.
Hide
Type "./guh --demo"
Enter
Sleep 500ms
Show

Wait+Screen /cartograph/
Sleep 2.2s
Right
Sleep 1.8s
Down@400ms 2
Sleep 900ms
Type "c"
Sleep 2.4s
Escape
Sleep 900ms
Type "o"
Sleep 1.8s
Down
Sleep 800ms
Enter
Wait+Screen /beacon/
Sleep 3.5s
