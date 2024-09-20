#!/bin/sh

hivechain generate \
  --fork-interval 6 \
  --tx-interval 1 \
  --length 500 \
  --outdir testdata \
  --lastfork london \
  --outputs accounts,genesis,chain,headstate,txinfo,headblock,headfcu,newpayload,forkenv
