rm -f example.db
sqlite3 example.db 'PRAGMA foreign_keys=on; CREATE TABLE loadbalance(id INTEGER PRIMARY KEY, usedMachines TEXT, totalMachinesCount INTEGER);'
sqlite3 example.db 'INSERT INTO loadbalance (usedMachines, totalMachinesCount) VALUES ("1", 1);'
sqlite3 example.db 'INSERT INTO loadbalance (usedMachines, totalMachinesCount) VALUES ("2", 2);'

sqlite3 example.db 'PRAGMA foreign_keys=on; CREATE TABLE machine(id INTEGER PRIMARY KEY, isWork INTEGER, loadbalance_id INTEGER,FOREIGN KEY (loadbalance_id) REFERENCES loadbalance(id));'
sqlite3 example.db 'INSERT INTO machine (isWork, loadbalance_id) VALUES (1,1);'

sqlite3 example.db 'INSERT INTO machine (isWork, loadbalance_id) VALUES (1,2);'