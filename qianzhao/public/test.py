def gen_file(path, size):
	file = open(path, "w")
	file.seek(1024 * 1024 * 1024 * size)
	file.write('\x00')
	file.close()

gen_file('./test_for_download.data', 1)
