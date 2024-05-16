build_test(){
    echo "Building test"
    go build tests/test.go
    mv test tests/test 
}

run_test(){
    echo "Running ./tests/test"
    build_test
    go build app/ccdocker.go
    ./ccdocker run container ./tests/test
}

run_ls(){
    echo "Running ls"
    go build app/ccdocker.go
    ./ccdocker run container ./tests/ls
}


export PATH=$PATH:/usr/local/go:/usr/local/go/bin


if [ ${#} -eq 0 ] ; then
    echo -e "\nUsage: ${0} [COMMANDS]\n\n Available commands:"
    cat `basename ${0}` | grep '()\s{' | while read COMMAND ; do echo " - ${COMMAND::-4}" ; done
else
    for COMMAND in "${@}" ; do "${COMMAND}" ; done
fi