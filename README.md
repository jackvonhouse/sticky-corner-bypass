# Sticky Corner Bypass

## Problem

Since Windows 10 (maybe earlier) when using few monitors it's impossible to move the cursor between screens because Microsoft make "sticky corner" (thanks 👏)

## Installing

```
go build -o cursor.exe ./main.go
```

Press ```Win + R```, then write ```shell:startup``` and paste ```cursor.exe``` there.