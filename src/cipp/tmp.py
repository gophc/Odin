# -*- coding: utf-8 -*-
from __future__ import print_function

import json

import redis
import os
import sys
import json

redis_uri = "redis://:wstest@localhost:6379"
rkey = 'tool_parser_x:req-1ec8b1c2-815b-019e-343b'


def fix_and_load(text):
	data = []
	fixed = text.replace("\n<|tool", "<|tool").replace("\"}]<|tool", "\"}}]<|tool").\
		replace("\"}]}]<|tool", "\"}]}}]<|tool").replace("\"}}<|tool", "\"}}]<|tool"). \
		replace("\"}}<|tool", "\"}}]<|tool").replace("\"}<|tool", "\"}}]<|tool")
	json_text = fixed.split('begin|>', 1)[1].split('<|tool', 1)[0]

	try:
		data = json.loads(json_text)
		assert isinstance(data, list), "json must as []"
		return data
	except Exception as ex:
		_LOG("json.loads len: %s, err: %r" % (rkey, ex))

	fix_arr = ['', '', '']
	for fix in [i for i in fix_arr if i]:
		pass

	return data


def action_write(name, arg):
	file_path, content = arg['file_path'], arg['content']
	with open(file_path, 'w') as wf:
		wf.write(content.encode('utf-8'))

	_LOG("write name: `%s`, file: `%s`, len: %d" % (name, file_path, len(content)))

T_MAP = {
	# {"name": "Write", "arguments": {"content": "", "file_path": "G:/b0paxxx/xxx.odin"}}
	'write': action_write,
	'unknown': lambda name, arg: _LOG("unknown name: `%s`, arg: `%s`" % (name, ','.join(arg.keys())))
}

def main():
	con = redis.Redis.from_url(redis_uri)
	text = con.get(rkey)
	_LOG("get key: %s, len: %d" % (rkey, len(text)))
	data = fix_and_load(text)
	for item in data:
		name = item['name'].lower() if 'name' in item and item['name'] else 'unknown'
		arguments = item['arguments'] if 'arguments' in item and isinstance(item['arguments'], dict) else {}
		fun = T_MAP.get(name, T_MAP['unknown'])
		fun(name, arguments)

def _LOG(msg, handle=None):
	print(msg, file=sys.stderr)
	if handle:
		handle.write(msg + '\n')
		handle.flush()


if __name__ == '__main__':
	_LOG("\n========== START ===========")
	main()
	_LOG("\n==========  END  ===========")
