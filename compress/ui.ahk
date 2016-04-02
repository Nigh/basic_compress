
#Include Gdip.ahk
#Include md5.ahk

SetWorkingDir %A_ScriptDir%
if(%0%<=0){
	Msgbox, 拖放文件至图标，Please.
	ExitApp
}
path=%1%
Array := Object()
Array.Insert(1,106,105,121,117,99,104,101,110,103,48,48,55,64,103,109,97,105,108,46,99,111,109)
str:=""
Loop, % Array.MaxIndex()
{
	str.=chr(Array[A_Index])
}
SplitPath, % path, OutFileName, OutDir, OutExt, OutNameNoExt, OutDrive

If !pToken := Gdip_Startup(){
	MsgBox, 48, gdiplus error!, Gdiplus failed to start. Please ensure you have gdiplus on your system
	ExitApp
}
OnExit, Exit
gui1W:=A_ScreenWidth
gui1H:=A_ScreenHeight
transparence:=170
Gui, 1:-Caption +hwndhgui1 +E0x80000 +AlwaysOnTop +Owner
Gui, 1:Show, x0 y0 w%gui1W% h%gui1H% NA
hbm1 := CreateDIBSection(gui1W, gui1H)
hdc1 := CreateCompatibleDC()
obm1 := SelectObject(hdc1, hbm1)
G1:=Gdip_GraphicsFromHDC(hdc1)
Gdip_SetSmoothingMode(G1, 4)

; option:=" cbbcccccc Bold r4 s20"

Gdip_FillRectangleWithColor(G1, transparence<<24, 0, 0, gui1W, gui1H)
rc:=display("Select Algorithm",{x:100,y:100}," s100 cbbcccccc Bold r4 w" A_ScreenWidth)
display("1. RLE compress",{y:100+rc.h+100}," s42 cbbcccccc Bold r4")
display("2. LZW compress(not finished yet)",""," s42 cbb777777 Bold r4")
display("3. Huffman compress",""," s42 cbbcccccc Bold r4")
display("4. RLE+Huffman compress")
display("5. DEFLATE compress(not finished yet)",""," s42 cbb777777 Bold r4")
display("6. LZMA compress(not finished yet)")
display(" ")
display("- Press ESC to exit")

displayAuthor(str," s18 c77cccccc r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Gdip_GraphicsClear(G1)
; SetFormat, Integer, HEX
Hotkey, 1 up, rle
Hotkey, 2 up, lzw
Hotkey, 3 up, Huffman
Hotkey, 4 up, HuffmanAndRLE
Hotkey, 5 up, lzw
Hotkey, 6 up, lzw
Return

hotkeyOff:
Hotkey, 1 up, off
Hotkey, 2 up, off
Hotkey, 3 up, off
Hotkey, 4 up, off
Hotkey, 5 up, off
Hotkey, 6 up, off
Return

lzw:
Rice:
DEFLATE:
LZMA:
Gosub, hotkeyOff
Return

HuffmanAndRLE:
Gosub, hotkeyOff
Gdip_FillRectangleWithColor(G1, transparence<<24, 0, 0, gui1W, gui1H)
_comp:=OutNameNoExt "_comp[RLE+Huffman].rh"
_tst:=OutNameNoExt "_check[RLE+Huffman]." OutExt
rc:=display("RLE+Huffman Compression",{x:100,y:100}," s100 cbbcccccc Bold r4 w" A_ScreenWidth)
display("Output Filename: "" " _comp " """,{y:100+rc.h+50}," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
IfExist, % _comp
FileDelete, % _comp
Sleep, 300

_var=compress.exe -input "%path%" -output temp.comp -d comp -m rle
nonblock_runwait(_var)
display("RLE Compress start...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 100
Loop
{
	if(is_nonblock_runwait_end()){
		display("RLE Compress finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 100
		Break
	}
}

_var=compress.exe -input temp.comp -output "%_comp%" -d comp -m huffman
nonblock_runwait(_var)
display("Huffman Compress start...")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 100
Loop
{
	if(is_nonblock_runwait_end()){
		display("Huffman Compress finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 100
		Break
	}
}
FileDelete, temp.comp

_var=compress.exe -input "%_comp%" -output temp.test -d dec -m huffman
nonblock_runwait(_var)
display("Self check start...")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Loop
{
	if(is_nonblock_runwait_end()){
		Break
	}
}
_var=compress.exe -input temp.test -output "%_tst%" -d dec -m rle
nonblock_runwait(_var)
Loop
{
	if(is_nonblock_runwait_end()){
		display("Self check finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
FileDelete, temp.test

md5_1:=MD5_File(path)
md5_2:=MD5_File(_tst)
FileGetSize, size_before, % path, K
FileGetSize, size_after, % _comp, K
FileDelete, % _tst
if(md5_1!=md5_2){
	display("Hash Check Error",""," s42 cbbff3322 Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display("Compression FAILED")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display(" ")
	display("- Press ESC to exit",""," s42 cbbcccccc Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Return
}
display("Hash check PASS",""," s42 cbb33ff22 Bold r4")
display("Compression SUCCESS")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300

display("Compression ratio:" Format("{:0.2f}", size_after/size_before*100) "%",""," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
_var=compress.exe -input "%path%" -d check
nonblock_runwait(_var)
display("Start generating checkfile...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Checkfile generated...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
display(" ")
display("- Press ESC to exit")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
FileMove, checkFile.chk, % OutDir, 1
FileMove, % _comp, % OutDir, 1
Return

Huffman:
Gosub, hotkeyOff
Gdip_FillRectangleWithColor(G1, transparence<<24, 0, 0, gui1W, gui1H)
rc:=display("Huffman Compression",{x:100,y:100}," s100 cbbcccccc Bold r4 w" A_ScreenWidth)
_comp:=OutNameNoExt "_comp[Huffman].hf"
_tst:=OutNameNoExt "_check[Huffman]." OutExt
display("Output Filename: "" " _comp " """,{y:100+rc.h+100}," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
IfExist, % _comp
FileDelete, % _comp
Sleep, 300

_var=compress.exe -input "%path%" -output "%_comp%" -d comp -m huffman
nonblock_runwait(_var)
display("Compress start...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Compress finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
_var=compress.exe -input "%_comp%" -output "%_tst%" -d dec -m huffman
nonblock_runwait(_var)
display("Self check start...")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Self check finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
md5_1:=MD5_File(path)
md5_2:=MD5_File(_tst)
FileGetSize, size_before, % path, K
FileGetSize, size_after, % _comp, K
FileDelete, % _tst
if(md5_1!=md5_2){
	display("Hash Check Error",""," s42 cbbff3322 Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display("Compression FAILED")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display(" ")
	display("- Press ESC to exit",""," s42 cbbcccccc Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Return
}
display("Hash check PASS",""," s42 cbb33ff22 Bold r4")
display("Compression SUCCESS")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300

display("Compression ratio:" Format("{:0.2f}", size_after/size_before*100) "%",""," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
_var=compress.exe -input "%path%" -d check
nonblock_runwait(_var)
display("Start generating checkfile...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Checkfile generated...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
display(" ")
display("- Press ESC to exit")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
FileMove, checkFile.chk, % OutDir, 1
FileMove, % _comp, % OutDir, 1
Return

rle:
Gosub, hotkeyOff
Gdip_FillRectangleWithColor(G1, transparence<<24, 0, 0, gui1W, gui1H)
rc:=display("RLE Compression",{x:100,y:100}," s100 cbbcccccc Bold r4 w" A_ScreenWidth)
_comp:=OutNameNoExt "_comp[RLE].rle"
_tst:=OutNameNoExt "_check[RLE]." OutExt
display("Output Filename: "" " _comp " """,{y:100+rc.h+100}," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
IfExist, % _comp
FileDelete, % _comp
Sleep, 300

_var=compress.exe -input "%path%" -output "%_comp%" -d comp -m rle
nonblock_runwait(_var)
display("Compress start...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Compress finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
_var=compress.exe -input "%_comp%" -output "%_tst%" -d dec -m rle
nonblock_runwait(_var)
display("Self check start...")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Self check finish...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
md5_1:=MD5_File(path)
md5_2:=MD5_File(_tst)
FileGetSize, size_before, % path, K
FileGetSize, size_after, % _comp, K
FileDelete, % _tst
if(md5_1!=md5_2){
	display("Hash Check Error",""," s42 cbbff3322 Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display("Compression FAILED")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Sleep, 1000
	display(" ")
	display("- Press ESC to exit",""," s42 cbbcccccc Bold r4")
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
	Return
}
display("Hash check PASS",""," s42 cbb33ff22 Bold r4")
display("Compression SUCCESS")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300

display("Compression ratio:" Format("{:0.2f}", size_after/size_before*100) "%",""," s42 cbb66aaff Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
_var=compress.exe -input "%path%" -d check
nonblock_runwait(_var)
display("Start generating checkfile...",""," s42 cbbcccccc Bold r4")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
Sleep, 300
Loop
{
	if(is_nonblock_runwait_end()){
		display("Checkfile generated...")
		UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
		Sleep, 300
		Break
	}
}
display(" ")
display("- Press ESC to exit")
UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
FileMove, checkFile.chk, % OutDir, 1
FileMove, % _comp, % OutDir, 1
Return

nonblock_runwait(var)
{
	global
	runwait_var:=var
	runwait_statu:=0
	SetTimer, runwait_process, -10
}

is_nonblock_runwait_end()
{
	global
	Return runwait_statu
}

runwait_process:
; MsgBox, % runwait_var
RunWait, % runwait_var, %A_ScriptDir%, Hide
runwait_statu:=1
Return

Esc::
Exit:
Gdip_Shutdown(pToken)
ExitApp

WM_LBUTTONUP()
{
	Gosub, Exit
}

rands(min=0,max=100)
{
	Random, var, % min, % max
	return var
}

randStr(length=22)
{
	str:=""
	loop, % length
	{
		str.=chr(rands(33,126))
	}
	return str
}

displayClear()
{
	global
	Gdip_GraphicsClear(G1)
	UpdateLayeredWindow(hgui1, hdc1, ,,,,255)
}

Gdip_FillRectangleWithColor(byref G, color, x, y, w, h)
{
	pBrush:=Gdip_BrushCreateSolid(color)
	Gdip_FillRectangle(G, pBrush, x, y, w, h)
	Gdip_DeleteBrush(pBrush)
}

displayAuthor(txt,opt="")
{
	global G1
	static option
	if(opt!=""){
		option:=opt
	}
	rc:=parseRC(Gdip_TextToGraphics(G1, txt, "x0 y0 w100p" option , "Arial",A_ScreenWidth,64,1))
	Gdip_TextToGraphics(G1, txt, "x" 0.98*A_ScreenWidth-rc.w " y" 0.98*A_ScreenHeight-rc.h " " option, "Arial",rc.w,rc.h)
	Return
}

display(txt,p="",opt="")
{
	global G1
	static option
	static pos:=Object()
	if(opt!=""){
		option:=opt
	}
	if(p.x){
		pos.x:=p.x
	}
	if(p.y){
		pos.y:=p.y
	}
	rc:=parseRC(Gdip_TextToGraphics(G1, txt, "x0 y0 w100p" option , "Arial",A_ScreenWidth,64,1))
	Gdip_TextToGraphics(G1, txt, "x" pos.x " y" pos.y " " option, "Arial",rc.w,rc.h)
	pos.y+=rc.h
	return rc
}

displayTxt(txt,index=0,opt="")
{
	global G1
	static option
	if(opt!=""){
		option:=opt
	}
	rc:=parseRC(Gdip_TextToGraphics(G1, txt, "x0 y0 w100p" option , "Arial",A_ScreenWidth,64,1))
	Gdip_TextToGraphics(G1, txt, "x" (A_ScreenWidth-rc.w)//2 " y" (A_ScreenHeight-rc.h-80)//2+index*rc.h option, "Arial",rc.w,rc.h)
	return rc
}

parseRC(rc)
{
	Loop, Parse, rc, |
	{
		if(A_Index=3)
			rc_w:=A_LoopField
		if(A_Index=4)
			rc_h:=A_LoopField
	}
	rc_w+=0
	rc_h+=0
	return {w:rc_w,h:rc_h}
}
