cat  > /tmp/test_list.txt <<EOF
region list
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
migration list

router list
network list
port list
network agent list
sg list
security group list
sg rule list
security group rule list
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
