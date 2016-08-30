./store \
-es-host  http://192.168.1.218:9200,http://192.168.1.218:9201 \
-geo-host  http://127.0.0.1:54321 \
-gtype  taobao_es \
-hbase-host  192.168.1.218 \
-hbase-port  2181 \
-mdb-put-host  192.168.1.199 \
-mdb-put-port  27017 \
-mdb-store-host  192.168.1.199 \
-mdb-store-port  27017 \
-nsq-host 127.0.0.1 \
-nsq-port 4150 \
-rKey zhejiang_goodsqueue_es  \
-table_prefixe zhejiang

