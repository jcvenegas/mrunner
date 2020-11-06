#!/bin/bash

find . -name 'fio-results-*' -exec echo {} \; -exec grep -P 'READ|WRITE|case' {} \;
