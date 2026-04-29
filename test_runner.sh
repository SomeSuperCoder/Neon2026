#!/bin/bash
cd /workspace
go test -v ./internal/consensus -run "TestRecordBlockProduction|TestBlockProductionCounter|TestResetAllBlockProductionCounters" -timeout 30s
