#!/bin/bash

gb build -ldflags='-s -w'

strip bin/*

upx bin/*


