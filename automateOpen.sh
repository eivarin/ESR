make
kitty --hold docker exec -it core vcmd -c /tmp/pycore.1/RP -- /bin/bash -i -c "clear && ./programs/OverlayRP/RP -n RP" &
sleep 1


for i in {1..10}
do
    kitty --hold docker exec -it core vcmd -c /tmp/pycore.1/O$i -- /bin/bash -i -c "clear && ./programs/OverlayNode/Node -n O$i" &
    sleep 0.5
    (bspc query -D --names | grep ^O$i$) || bspc monitor -a O$i
    bspc node -d O$i
    sleep 0.5
done

for i in {1..3}
do
    kitty --hold docker exec -it core vcmd -c /tmp/pycore.1/S$i -- /bin/bash -i -c "clear && ./programs/OverlayServer/Server -n S$i" &
    sleep 0.5
    (bspc query -D --names | grep ^S$i$) || bspc monitor -a S$i
    bspc node -d S$i
    sleep 0.5
done

for i in {1..2}
do
    kitty --hold docker exec -it core vcmd -c /tmp/pycore.1/C$i -- /bin/bash -i -c "clear && ./programs/OverlayClient/Client -n C$i" &
    sleep 0.5
    (bspc query -D --names | grep ^C$i$) || bspc monitor -a C$i
    bspc node -d C$i
    sleep 0.5
done