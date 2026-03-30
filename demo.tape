Output kepfi-demo.gif

#Set Theme "Catppuccin Frappe"
Set FontSize 16
Set Width 1200
Set Height 600

Hide
  Type "export PATH=$PATH:$(pwd)/../bin"
  Enter
  Type "dd if=/dev/zero of=zero.bin bs=1M count=50"
  Enter
  Type "kepfi -pa -f && touch file.txt && mkdir testfolder && clear"
  Enter
Show


Type "ls"
Enter
Sleep 3s

Type "kepfi file.txt zero.bin testfolder"
Enter
Sleep 3s

Type "kepfi -l"
Enter
Sleep 3s

Type "kepfi -r file.txt"
Enter
Sleep 3s

Type "ls"
Enter  
Sleep 3s

Hide
  Type "rm -r file.txt zero.bin testfolder && clear"
  Enter
Show
