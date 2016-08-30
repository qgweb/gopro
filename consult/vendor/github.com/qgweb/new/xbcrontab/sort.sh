#!/bin/bash
if [ ! -e "$1" ];then
exit -1
fi
lines=$(wc -l $1 | sed 's/ .*//g')
num=20
if [ $lines -eq 0 ]; then
exit 0
fi
if [ $lines -lt $num ]; then
num=$lines
fi
lines_per_file=`expr $lines / $num`
split -d -l $lines_per_file $1 __part_${1##*/}
for file in __part_*
do
{
  sort $file > sort_$file
} &
done
wait
sort -smu sort_* > $1
rm -f __part_*
rm -f sort_*