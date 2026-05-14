# -*- coding: UTF-8 -*-
from __future__ import print_function

import datetime
import re
import subprocess
from multiprocessing import Pool
import os
import time

wait_map, results_map = {}, {}


# region --- do multiprocessing ---

def do_done(result):
	key = result['key']
	wait_map[key] = None
	results_map[key] = result
	print('task done: %s' % (key,))


def do_init():
	pass


def do_task(args):
	args = args.encode('utf-8') if isinstance(args, unicode) else str(args)
	pid, start_time = os.getpid(), time.time()
	key, cmd = args.split('#')[0], args.split('#')[1]
	print('pid: ' + str(pid) + ' `' + key + '`, cmd: ' + cmd)
	try:
		cmd_ = cmd.split(' ')
		process = subprocess.Popen(
			cmd_,
			stdout=subprocess.PIPE,
			stderr=subprocess.PIPE
		)
		stdout_, stderr_ = process.communicate()
		stdout = stdout_.decode('utf-8', errors='ignore') if stdout_ else ""
		stderr = stderr_.decode('utf-8', errors='ignore') if stderr_ else ""

		print('pid: %d `%s`' % (pid, key) +
			  ', out [%d]: %s' % (len(stdout), stdout.strip().split('\n')[-1]) +
			  ', err [%d]: %s' % (len(stderr), stderr.strip().split('\n')[-1]))
		end_time = time.time()
		result = {
			'pid': pid,
			'key': key,
			'code': process.returncode,
			'stdout': stdout,
			'stderr': stderr,
			'time_taken': end_time - start_time,
		}
		return result
	except Exception as e:
		print('pid: ' + str(pid) + ' `' + key + '`, err: ' + repr(e))
		end_time = time.time()
		result = {
			'pid': pid,
			'key': key,
			'code': -1,
			'stdout': '',
			'stderr': repr(e),
			'time_taken': end_time - start_time,
		}
		return result


# endregion

def main(
		dns_text, pool_size=18, action='ping+default',
		url='https://github.com/status', host='github.com', port=443,
		args='-v --connect-timeout 20 -m 30'
):
	dns_text_ = _pre_fix_text(dns_text)
	dns_set_ = set([l.strip() for l in dns_text_.split('\n') if l.strip()])
	dns_arr_ = [_pre_fix_line(l) for l in dns_set_ if _pre_fix_line(l)]
	dns_arr_.sort(key=lambda o: int(o[3]))

	dns_arr, ip_map = [], {}
	for idx, item in enumerate(dns_arr_, 0):
		if item[1] in ip_map: continue
		ip_map[item[1]] = item
		dns_arr.append(item)

	log('got ipv4 len: %d' % (len(dns_arr),))

	for item in dns_arr: log(' '.join(item), 'info' if 'ping' not in action else 'debug')

	tasks = {}
	for item in dns_arr:
		# noinspection PyUnresolvedReferences
		curl_cmd = 'curl %s --resolve "%s:%d:%s" %s' % (args.strip(), host.strip(), port, item[1], url.strip())
		log(curl_cmd, 'info' if 'ping' not in action else 'debug')
		tasks[item[1]] = '%s#%s' % (item[1], curl_cmd)

	if 'ping' not in action:
		return

	pool_size = int(pool_size) if int(pool_size) > 1 else 1
	if len(tasks) <= 4 or pool_size <= 2:
		log(('pool_size: %d tasks: %d' % (pool_size, len(tasks))) + 'tasks:' +
			','.join(["'%s'" % (_c,) for _c in tasks.keys()]) if tasks else '[]')
		results = []
		for code, args in tasks.items():
			tmp = do_task(args)
			results.append(tmp)
		log('done tasks: %d ' % (len(results),))
		_print_results(results)
		return

	log("Starting %d processes to fetch %d URLs..." % (pool_size, len(tasks)))
	pool = Pool(pool_size, initializer=do_init, initargs=())
	for code, curl_cmd in tasks.items():
		wait_map[code] = pool.apply_async(do_task, (curl_cmd,),
										  callback=do_done)

	n, not_done = 0, {}
	while n <= 10:
		n += 1
		not_done = {k: v for k, v in wait_map.items() if v is not None}
		print('not done tasks: ' + ','.join(["'%s'" % (_c,) for _c in not_done.keys()])) \
			if not_done and len(not_done) <= 10 else None
		for c, r in not_done.items():
			# noinspection PyBroadException
			try:
				r.get(timeout=1)
				wait_map[c] = None
				print('task done: %s' % (c,))
			except Exception:
				pass

	pool.close()
	pool.join()

	log("All tasks completed.")

	_print_results(results_map.values())


# region --- help func ---

def _print_results(results):
	log("Received %d results" % (len(results),))
	for res in results:
		log("PID: {}, KEY: {}, Code: {}, Time: {:.2f}s".format(
			res['pid'], res['key'], res['code'], res['time_taken']
		), 'info')
		print(res['stdout'] if res['stdout'] else res['stderr'])


def _pre_fix_text(dns_text):
	dns_text = dns_text if isinstance(dns_text, unicode) else dns_text.decode('utf-8')
	dns_text = dns_text.replace(u'海外', '').replace('\r', '\n').replace('\t', ' ')
	dns_text = dns_text.replace('\n\n', '\n')
	while '  ' in dns_text:
		dns_text.replace('  ', ' ')
	while '\n\n' in dns_text:
		dns_text.replace('\n\n', '\n')

	return dns_text


def _pre_fix_line(line):
	if 'ms' not in line:
		return []
	line_ = line.rsplit('ms', 1)[0]
	arr = line_.rsplit(' ', 1)
	arr_, delay = arr[0].split(' ', 2), arr[-1]
	_is_ipv4 = lambda s: re.match(r"^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$", s)
	assert len(arr_) == 3 and delay.isdigit(), "must delay.isdigit() but %s in %s" % (delay, line)
	assert _is_ipv4(arr_[1]), "must ipv4 but %s in %s" % (arr_[1], line)
	return [arr_[0], arr_[1], arr_[2], delay]


# endregion

LOG_LEVEL = 'info'


# region --- LOG HELPER ---

def T(tag='info'):
	time_str = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]
	return '%s [%s] ' % (time_str, tag.upper())


def log(msg, tag='info', handle=None):
	if not msg: return
	level_map = {'debug': 10, 'info': 20, 'warn': 30, 'error': 40, 'fault': 50}
	level, level_ = level_map.get((tag or 'info').lower(), 20), level_map.get((LOG_LEVEL or '').lower(), 10)
	if level < level_: return
	_msg = T(tag or 'info') + msg.replace("\n", r"\n")
	print(_msg)
	handle = handle if handle else getattr(log, 'handle', None)
	if handle:
		handle.write(_msg + '\n')
		handle.flush()


# endregion

DEFAULT_DNS_TEXT = u'''
海外
香港	20.205.243.166	新加坡/新加坡/微软公司	30ms	【高防CDN】死扛攻击
海外
日本	20.27.177.113	日本/东京都/东京/微软公司	1ms	网速测试
海外
香港	20.205.243.166	新加坡/新加坡/微软公司	36ms	公共DNS大全
海外
日本	20.27.177.113	日本/东京都/东京/微软公司	3ms	赞助QQ：123480129
海外
香港	20.205.243.166	新加坡/新加坡/微软公司	34ms	赞助QQ：123480129
海外
香港	20.205.243.166	新加坡/新加坡/微软公司	35ms	赞助QQ：123480129
海外
泰国	20.205.243.166	新加坡/新加坡/微软公司	28ms	赞助QQ：123480129
海外
澳大利亚	4.237.22.38	澳大利亚/新南威尔士州/悉尼/微软公司	1ms	赞助QQ：123480129
海外
台湾台北	20.27.177.113	日本/东京都/东京/微软公司	45ms	赞助QQ：123480129
海外
日本	20.27.177.113	日本/东京都/东京/微软公司	2ms	赞助QQ：123480129
海外
越南	20.205.243.166	新加坡/新加坡/微软公司	59ms	赞助QQ：123480129
海外
澳大利亚	4.237.22.38	澳大利亚/新南威尔士州/悉尼/微软公司	3ms	赞助QQ：123480129
海外
日本	20.27.177.113	日本/东京都/东京/微软公司	1ms	赞助QQ：123480129
海外
荷兰	140.82.121.4	德国/黑森州/美因河畔法兰克福/GitHub, Inc.	7ms	赞助QQ：123480129
海外
德国	140.82.121.3	德国/黑森州/美因河畔法兰克福/GitHub, Inc.	17ms	赞助QQ：123480129
海外
立陶宛	140.82.121.3	德国/黑森州/美因河畔法兰克福/GitHub, Inc.	30ms	赞助QQ：123480129
海外
美国	140.82.114.3	美国/弗吉尼亚州/阿什本/GitHub, Inc.	53ms	赞助QQ：123480129
海外
美国	140.82.116.4	美国/华盛顿州/西雅图/GitHub, Inc.	29ms	赞助QQ：123480129
海外
印度	20.207.73.82	印度/马哈拉施特拉邦/浦那/微软公司	4ms	赞助QQ：123480129
海外
加拿大	140.82.113.3	美国/弗吉尼亚州/阿什本/GitHub, Inc.	18ms	赞助QQ：123480129
海外
摩尔多瓦	140.82.121.4	德国/黑森州/美因河畔法兰克福/GitHub, Inc.	37ms	赞助QQ：123480129
海外
印度	20.207.73.82	印度/马哈拉施特拉邦/浦那/微软公司	4ms	赞助QQ：123480129
海外
智利	4.228.31.150	巴西/圣保罗州/圣保罗/微软公司	49ms	赞助QQ：123480129
海外
尼日利亚	140.82.121.4	德国/黑森州/美因河畔法兰克福/GitHub, Inc.	116ms	赞助QQ：123480129
海外
美国	140.82.112.4	美国/弗吉尼亚州/阿什本/GitHub, Inc.	58ms	赞助QQ：123480129
海外
美国	20.205.243.166	新加坡/新加坡/微软公司	172ms	赞助QQ：123480129
'''

if __name__ == '__main__':
	main(DEFAULT_DNS_TEXT)
