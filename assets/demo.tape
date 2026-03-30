Output kepfi-demo.gif

#Set Theme "Catppuccin Frappe"
Set FontSize 20
Set Width 1200
Set Height 600
Set WindowBar Colorful

Hide
  Type "export PATH=$PATH:$(pwd)/../bin"
  Enter
  Type "mkdir -p /tmp/demo && cd /tmp/demo && kepfi -pa -f && touch file.txt && mkdir testfolder"
  Enter
  Type "dd if=/dev/zero of=zero.bin bs=1M count=50 && clear"
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
  Type "cd .. && rm -r demo && clear"
  Enter
Show
