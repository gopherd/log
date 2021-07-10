#!/bin/bash

echo "// Auto-generated by github.com/gopherd/log/genlintfuncs.sh, DON'T EDIT IT!"
echo "package analyzer"
echo
echo "var allFuncs = []string{"
grep -h " \*Recorder {" *.go | grep '^func [A-Z]' | sed "s/(/\ /g" | awk '{printf ("\t\"%s\",\n",$2)}' | sort
echo "}"
echo
echo "var allMethods = []string{"
grep -h " \*Recorder {" *.go | grep '^func (.*) [A-Z]' | sed "s/func\ (//g" | sed "s/(/\ /g" | sed "s/)//g" | awk '{printf ("\t\"%s.%s\",\n",$2,$3)}' | sort
echo "}"
