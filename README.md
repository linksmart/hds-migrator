# Historical Datastore (HDS) Migration tool
Historical Datastore (HDS) migration tool is a command line tool used to migrate data from one HDS to another.

## Compile and Run from Source
In order to compile, you need [go](https://golang.org/dl/) in your system. If you have it, please follow the following instructions:
1. Clone the repository
````shell script
git clone https://github.com/linksmart/hds-migrator.git
````
2. cd to the directory and build the code
````shell script
go build -o hds-migrator
````
3. Execute the command
````shell script
hds-migrator http://source.example.com http://dest.example.ceom
````

## Contributing
Contributions are welcome. 

Please fork, make your changes, and submit a pull request. For major changes, please open an issue first and discuss it with the other authors.
