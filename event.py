import json

msg = {
    'name' : 'John Doe',
    'department' : 'Marketing',
    'place' : 'Remote'
}

json_string = json.dumps(msg)
print(json_string)