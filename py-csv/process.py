#!/anaconda3/bin/python

import csv
import sys


def main():
    filename = "data.csv"
    if len(sys.argv) > 1:
        filename = sys.argv[1]

    w = csv.writer(sys.stdout)
    # w.writerow(["col1,col1"])

    with open(filename, newline='') as f:
        reader = csv.DictReader(f)
        for row in reader:
            w.writerow(doSomething(row).values())
                
def doSomething(row):
    # row['last'] = "new"
    return row

if __name__== "__main__":
  main()

