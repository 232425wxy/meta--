test:
	## 如果不存在0号节点的创世文件，则从头创建文件并开始
	@if ! [ -f /root/meta--/node0/genesis.json ]; then echo "hei"; fi
.PHONY: test