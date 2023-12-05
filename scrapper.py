import json

root = json.load(open('Final.json', 'r'))
devices = root['devices']
links = root['links']

name_to_id = {}
for device in devices:
    name_to_id[device['@name']] = device['@id']

id_ips = {}
for link in links:
    if link['@node1'] not in id_ips:
        id_ips[link['@node1']] = set()
    if '@ip4' in link['iface1']:
        id_ips[link['@node1']].add(link['iface1']['@ip4'])
    if link['@node2'] not in id_ips:
        id_ips[link['@node2']] = set()
    if '@ip4' in link['iface2']:
        id_ips[link['@node2']].add(link['iface2']['@ip4'])


name_to_ip = {name: list(id_ips[name_to_id[name]])[0] for name in name_to_id}

mapping = {
    'RP': [],
    'O1': ['RP'],
    'O2': ['O1', 'RP'],
    'O3': ['O2', 'RP'],
    'O4': ['O1', 'O3', 'RP'],
    'O5': ['O4'],
    'O6': ['O1', 'O5'],
    'O7': ['O1', 'O2', 'O6'],
    'O8': ['O2', 'O3'],
    'O9': ['O8'],
    'O10': ['O3', 'O4', 'O9'],
    'C1': ['O7'],
    'C2': ['O7'],
    'C3': ['O7'],
    'C4': ['O7'],
    'C5': ['O7'],
    'C6': ['O7'],
    'C7': ['O2'],
    'C8': ['O2'],
    'S1': ['O2'],
    'C9': ['O8'],
    'S2': ['O9'],
    'C10': ['O9'],
    'C11': ['O9'],
    'C12': ['O9'],
    'C13': ['O10'],
    'C14': ['O10'],
    'S3': ['O4'],
    'C15': ['O4'],
    'C16': ['O4'],
    'C17': ['O4'],
    'C18': ['O5'],
    'C19': ['O5'],
    'C20': ['O6'],
    'C21': ['O6'],
    'C22': ['O6'],
    'C23': ['O6']
}
new_mapping = {}
for name in mapping:
    new_mapping[name] = []
    for n in mapping[name]:
        new_mapping[name].append(name_to_ip[n])

for name in new_mapping:
    with open(f"./static/configs/{name}.conf", 'w') as f:
        f.write(f"{name_to_ip[name]}\n")
        for ip in new_mapping[name]:
            f.write(f"{ip}\n")

for name in ["S1", "S2", "S3"]:
    with open(f"./static/configs/{name}.conf", 'a') as f:
        f.write(f"./static/manifests/{name}.manifest\n")