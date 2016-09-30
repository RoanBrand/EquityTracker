from googlefinance import getQuotes
import json
import sys

print json.dumps(getQuotes(sys.argv[1]))
