echo "start qlc node"

if [ ! -n "$seed" ]; then
    # node mode
    exec ./gqlc
else
    # account mode
    exec ./gqlc  --seed=${seed} 
fi
