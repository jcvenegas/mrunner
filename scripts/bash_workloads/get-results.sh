#!/bin/bash

find . -name 'log-*' -exec echo {} \; -exec grep -P 'READ|WRITE|case' {} \;
