cat  > /tmp/test_list.txt <<EOF
service list
endpoint list
project list
user list

image list

volume list
volume type list

server list
flavor list
compute service list
hypervisor list
aggregate list
az list
az list --tree

router list
network list
port list
EOF

while read line
do
    if [[ "$line" == "" ]]; then
        continue
    fi
    echo "#### $line ####"
    go run cmd/skyman.go $line || break
done < /tmp/test_list.txt


rm -f /tmp/test_list.txt
