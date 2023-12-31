// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package rowexec

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	"github.com/opentracing/opentracing-go"
	"github.com/znbasedb/errors"
	"github.com/znbasedb/znbase/pkg/sql/execinfra"
	"github.com/znbasedb/znbase/pkg/sql/execinfrapb"
	"github.com/znbasedb/znbase/pkg/sql/sem/tree"
	"github.com/znbasedb/znbase/pkg/sql/sqlbase"
	"github.com/znbasedb/znbase/pkg/sql/types"
	"github.com/znbasedb/znbase/pkg/util/humanizeutil"
	"github.com/znbasedb/znbase/pkg/util/log"
	"github.com/znbasedb/znbase/pkg/util/mon"
	"github.com/znbasedb/znbase/pkg/util/stringarena"
	"github.com/znbasedb/znbase/pkg/util/tracing"
)

type aggregateFuncs []tree.AggregateFunc

func (af aggregateFuncs) close(ctx context.Context) {
	for _, f := range af {
		f.Close(ctx)
	}
}

type fpgaAggregator struct {
	data        []byte   // data block (64 rows / block)
	crt_rows    uint32   // current inserted number of rows
	byte_schema []int32  // For offset
	row_length  int32    // Single row size
	colId       []uint32 // Cols involved
	colType     []string

	results []byte
	res_ptr int

	row_size  int
	groupCols []uint32

	accu_size uint64 // Total size (byte)

	block_capacity int
	datumAlloc     sqlbase.DatumAlloc

	tres int

	// fpgahandle C.FPGAHandle
}

func initFpgaAggregator(gc []uint32, rs int, origin_schema []types.T) *fpgaAggregator {
	ret := &fpgaAggregator{}
	ret.data = make([]byte, 0)
	ret.crt_rows = 0
	ret.accu_size = 0
	ret.row_size = rs
	ret.groupCols = gc
	ret.block_capacity = 64

	ret.byte_schema = make([]int32, 0)
	ret.row_length = 0
	for _, col := range origin_schema {
		ret.byte_schema = append(ret.byte_schema, col.Width())
		ret.row_length += col.Width()
	}

	ret.results = make([]byte, 0)
	ret.res_ptr = 0

	ret.tres = 0

	// ret.fpgaHandle = C.FPGA_open()

	return ret
}

func (fa *fpgaAggregator) transS2Enum(t string) /*C.AggFunc*/ {
	/*
		switch (t) {
			case "count": {
				return C.AggFunc_count
			}
			case "sum": {
				return C.AggFunc_sum
			}
			case "max": {
				return C.AggFunc_max
			}
			case "min": {
				return C.AggFunc_min
			}
			case "avg": {
				return C.AggFunc_avg
			}
			default: {
				return C.AggFunc_noAgg
			}
		}

	*/
	fmt.Println("TransS2Enum")
}

func (fa *fpgaAggregator) recvResult() {
	/*if len(fa.data) != 0 {
		fa.submitRows()
	}
	result_length := 0
	for i, v := range fa.colType {
		switch v {
		case "count", "sum", "avg":
			{
				result_length += 8
			}
		case "min", "max":
			{
				result_length += int(fa.byte_schema[fa.colId[i]])
			}
		default:
			{
				result_length += 0
			}
		}
	}
	fa.results = make([]byte, result_length)*/
	/*
		var tempSize C.uint64_t = (C.uint64_t)(uint64(len(fa.result_length)))
		nRet := C.FPGA_next_result(handleCall, (*C.void)(fa.result), &tempSize)
		if (C.EFPGA_OK != nRet && C.EFPGA_NEXT_DATA != nRet) {
			fmt.Println("FPGA_next_result err :", nRet)

		}
	*/
	fmt.Println("RecvResult")
}

func (fa *fpgaAggregator) getResult() int {
	fmt.Println("GetResult")

	/*if len(fa.results) == 0 {
		fa.recvResult()
	}

	var start int = 0
	var s int = 0
	for i := 0; i < fa.res_ptr; i += 1 {
		switch fa.colType[i] {
		case "count", "sum", "avg":
			{
				start += 8
			}
		case "min", "max":
			{
				start += int(fa.byte_schema[fa.colId[i]])
			}
		default:
			{
				start += 0
			}
		}
	}
	switch fa.colType[fa.res_ptr] {
	case "count", "sum", "avg":
		{
			s = 8
		}
	case "min", "max":
		{
			s = int(fa.byte_schema[fa.colId[fa.res_ptr]])
		}
	default:
		{
			s = 0
		}
	}

	fa.res_ptr += 1

	var temp []byte = fa.results[start:s]
	bits := binary.LittleEndian.Uint64(temp)
	return math.Float64frombits(bits)*/
	return fa.tres
}

func (fa *fpgaAggregator) submitConfig() {
	fmt.Println("SubmitConfig")

	/*
		pConfigInfo *C.ConfgInfo = (*C.ConfigInfo)C.malloc(C.sizeof(C.ConfigInfo) + len(fa.colId) * C.sizeof(C.ColInfo))
		pConfigInfo.total = fa.rs * fa.row_length
		pConfigInfo.lrow = fa.rs
		pConfigInfo.count = len(colId)

		for i, gc := range fa.groupCols {
			pConfigInfo.info[i].offset = off
			pconfigInfo.info[i].size = fa.byte_schema[i]
			pconfigInfo.info[i].agginfo.enGroup = true
			pconfigInfo.info[i].agginfo.aggFunc = C.AggFunc_noAgg
			pconfigInfo.info[i].sortinfo.asc = 1
		}

		for i, id := range fa.colId {
			off := 0
			for j := 0; j < i; j += 1 {
				off += fa.byte_schema[j]
			}
			pConfigInfo.info[i].offset = off
			pconfigInfo.info[i].size = fa.byte_schema[i]
			pconfigInfo.info[i].agginfo.enGroup = false
			pconfigInfo.info[i].agginfo.aggFunc = fa.transS2Enum(fa.aggType[i])
		}
		var sort_info_size C.int = C.sizeof(C.ConfigInfo) + len(fa.colId) * C.sizeof(C.ColInfo)
		var handleCall C.CallHandle = C.FPGA_call(fa.handleFpga, C.OPGROUP, pConfigInfo, sort_info_size)
		if !handleCall {
			fmt.Println("FPGA_call faild")

		}
	*/
}

func (fa *fpgaAggregator) submitRows() {
	fmt.Println("SubmitRows")

	/*
		var pstDataInfo *C.DataInfo
		pstDataInfo = (*C.DataInfo)C.malloc(C.sizeof(C.DataInfo) + 1 * C.sizeof(C.DataNode))
		pstDataInfo.count = 1
		pstDataInfo.data[0].data = (*C.void)(fa.data)
		pstDataInfo.data[0].len = len(fa.data)

		nRet := C.FPGA_next_data(fa.handleCall, pstDataInfo)
		if (C.EFPGA_OK != nRet && C.EFPGA_NEXT_DATA != nRet) {
			fmt.Println("FPGA_next_data err :", nRet)

		}
	*/

	fa.data = make([]byte, 0)
	fa.crt_rows = 0
}

func (fa *fpgaAggregator) append(row sqlbase.EncDatumRow, inputTypes []types.T) {
	fa.tres += 1
	fmt.Println("Append row: ", fa.tres)

	for i, col := range row {
		if err := col.EnsureDecoded(&inputTypes[i], &fa.datumAlloc); err != nil {
			fmt.Println("Error")
		}
		// 转化成byte append进去
		fa.data = append(fa.data, []byte("")...)
		fa.crt_rows += 1
	}
	if fa.crt_rows == uint32(fa.block_capacity) {
		fa.submitRows()
	}
}

func (fa *fpgaAggregator) appendCol(col []uint32, t []string) int {
	fmt.Println("Append colId :", col, "type :", t[0])

	fa.colId = append(fa.colId, col...)
	if len(t) == 1 {
		for i := 0; i < len(col); i += 1 {
			fa.colType = append(fa.colType, t...)
		}
	} else {
		fa.colType = append(fa.colType, t...)
	}
	return len(fa.colId)
}

// aggregatorBase is the foundation of the processor core type that does
// "aggregation" in the SQL sense. It groups rows and computes an aggregate for
// each group. The group is configured using the group key and the aggregator
// can be configured with one or more aggregation functions, as defined in the
// AggregatorSpec_Func enum.
//
// aggregatorBase's output schema is comprised of what is specified by the
// accompanying SELECT expressions.
type aggregatorBase struct {
	execinfra.ProcessorBase

	// runningState represents the state of the aggregator. This is in addition to
	// ProcessorBase.State - the runningState is only relevant when
	// ProcessorBase.State == StateRunning.
	runningState aggregatorState
	input        execinfra.RowSource
	inputDone    bool
	inputTypes   []types.T
	funcs        []*aggregateFuncHolder
	outputTypes  []types.T
	datumAlloc   sqlbase.DatumAlloc
	rowAlloc     sqlbase.EncDatumRowAlloc

	bucketsAcc  mon.BoundAccount
	aggFuncsAcc mon.BoundAccount

	// isScalar can only be set if there are no groupCols, and it means that we
	// will generate a result row even if there are no input rows. Used for
	// queries like SELECT MAX(n) FROM t.
	isScalar         bool
	groupCols        []uint32
	orderedGroupCols []uint32
	aggregations     []execinfrapb.AggregatorSpec_Aggregation

	lastOrdGroupCols sqlbase.EncDatumRow
	arena            stringarena.Arena
	row              sqlbase.EncDatumRow
	scratch          []byte

	cancelChecker *sqlbase.CancelChecker

	fpgaAgg *fpgaAggregator // Append fpga
	// useFpga bool            // Trigger
}

func canUseFpga(y string) bool {
	return y == "count" || y == "sum" || y == "avg"
}

// init initializes the aggregatorBase.
//
// trailingMetaCallback is passed as part of ProcStateOpts; the inputs to drain
// are in aggregatorBase.
func (ag *aggregatorBase) init(
	self execinfra.RowSource,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
	trailingMetaCallback func(context.Context) []execinfrapb.ProducerMetadata,
) error {
	ctx := flowCtx.EvalCtx.Ctx()
	memMonitor := execinfra.NewMonitor(ctx, flowCtx.EvalCtx.Mon, "aggregator-mem")
	if sp := opentracing.SpanFromContext(ctx); sp != nil && tracing.IsRecording(sp) {
		input = newInputStatCollector(input)
		ag.FinishTrace = ag.outputStatsToTrace
	}
	ag.input = input
	ag.isScalar = spec.IsScalar()
	ag.groupCols = spec.GroupCols
	ag.orderedGroupCols = spec.OrderedGroupCols
	ag.aggregations = spec.Aggregations
	ag.funcs = make([]*aggregateFuncHolder, len(spec.Aggregations))
	ag.outputTypes = make([]types.T, len(spec.Aggregations))
	ag.row = make(sqlbase.EncDatumRow, len(spec.Aggregations))
	ag.bucketsAcc = memMonitor.MakeBoundAccount()
	ag.arena = stringarena.Make(&ag.bucketsAcc)
	ag.aggFuncsAcc = memMonitor.MakeBoundAccount()
	// Loop over the select expressions and extract any aggregate functions --
	// non-aggregation functions are replaced with parser.NewIdentAggregate,
	// (which just returns the last value added to them for a bucket) to provide
	// grouped-by values for each bucket.  ag.funcs is updated to contain all
	// the functions which need to be fed values.
	ag.inputTypes = input.OutputTypes()

	var count int = 0
	/*temp_input := input
	// 计数
	for {
		r, _ := temp_input.Next()
		if r == nil {
			break
		} else {
			count += 1
		}
	}*/

	ag.fpgaAgg = initFpgaAggregator(ag.groupCols, count, ag.inputTypes) // Append fpga
	// ag.useFpga = true
	for i, aggInfo := range spec.Aggregations {
		if aggInfo.FilterColIdx != nil {
			col := *aggInfo.FilterColIdx
			if col >= uint32(len(ag.inputTypes)) {
				return errors.Errorf("FilterColIdx out of range (%d)", col)
			}
			t := ag.inputTypes[col].Family()
			if t != types.BoolFamily && t != types.UnknownFamily {
				return errors.Errorf(
					"filter column %d must be of boolean type, not %s", *aggInfo.FilterColIdx, t,
				)
			}
		}

		argTypes := make([]types.T, len(aggInfo.ColIdx)+len(aggInfo.Arguments))
		for j, c := range aggInfo.ColIdx {
			if c >= uint32(len(ag.inputTypes)) {
				return errors.Errorf("ColIdx out of range (%d)", aggInfo.ColIdx)
			}
			argTypes[j] = ag.inputTypes[c]
		}

		arguments := make(tree.Datums, len(aggInfo.Arguments))
		for j, argument := range aggInfo.Arguments {
			h := execinfra.ExprHelper{}
			// Pass nil types and row - there are no variables in these expressions.
			if err := h.Init(argument, nil /* types */, flowCtx.EvalCtx); err != nil {
				return errors.Wrapf(err, "%s", argument)
			}
			d, err := h.Eval(nil /* row */)
			if err != nil {
				return errors.Wrapf(err, "%s", argument)
			}
			argTypes[len(aggInfo.ColIdx)+j] = *d.ResolvedType()
			if err != nil {
				return errors.Wrapf(err, "%s", argument)
			}
			arguments[j] = d
		}

		aggConstructor, retType, err := execinfrapb.GetAggregateInfo(aggInfo.Func, argTypes...)
		if err != nil {
			return err
		}

		fn_type := strings.ToLower(aggInfo.Func.String())
		if canUseFpga(fn_type) && retType.Equal(*types.Int4) {
			fmt.Println("Append agg  type: ", fn_type)
			ag.fpgaAgg.appendCol(aggInfo.ColIdx, []string{fn_type})
		} else {
			ag.funcs[i] = ag.newAggregateFuncHolder(aggConstructor, arguments)
			if aggInfo.Distinct {
				ag.funcs[i].seen = make(map[string]struct{})
			}
		}

		ag.outputTypes[i] = *retType
	}

	return ag.ProcessorBase.Init(
		self, post, ag.outputTypes, flowCtx, processorID, output, memMonitor,
		execinfra.ProcStateOpts{
			InputsToDrain:        []execinfra.RowSource{ag.input},
			TrailingMetaCallback: trailingMetaCallback,
		},
	)
}

var _ execinfrapb.DistSQLSpanStats = &AggregatorStats{}

const aggregatorTagPrefix = "aggregator."

// Stats implements the SpanStats interface.
func (as *AggregatorStats) Stats() map[string]string {
	inputStatsMap := as.InputStats.Stats(aggregatorTagPrefix)
	inputStatsMap[aggregatorTagPrefix+MaxMemoryTagSuffix] = humanizeutil.IBytes(as.MaxAllocatedMem)
	return inputStatsMap
}

// StatsForQueryPlan implements the DistSQLSpanStats interface.
func (as *AggregatorStats) StatsForQueryPlan() []string {
	stats := as.InputStats.StatsForQueryPlan("" /* prefix */)

	if as.MaxAllocatedMem != 0 {
		stats = append(stats,
			fmt.Sprintf("%s: %s", MaxMemoryQueryPlanSuffix, humanizeutil.IBytes(as.MaxAllocatedMem)))
	}

	return stats
}

func (ag *aggregatorBase) outputStatsToTrace() {
	is, ok := getInputStats(ag.FlowCtx, ag.input)
	if !ok {
		return
	}
	if sp := opentracing.SpanFromContext(ag.Ctx); sp != nil {
		tracing.SetSpanStats(
			sp,
			&AggregatorStats{
				InputStats:      is,
				MaxAllocatedMem: ag.MemMonitor.MaximumBytes(),
			},
		)
	}
}

// ChildCount is part of the execinfra.OpNode interface.
func (ag *aggregatorBase) ChildCount(verbose bool) int {
	if _, ok := ag.input.(execinfra.OpNode); ok {
		return 1
	}
	return 0
}

// Child is part of the execinfra.OpNode interface.
func (ag *aggregatorBase) Child(nth int, verbose bool) execinfra.OpNode {
	if nth == 0 {
		if n, ok := ag.input.(execinfra.OpNode); ok {
			return n
		}
		panic("input to aggregatorBase is not an execinfra.OpNode")
	}
	panic(fmt.Sprintf("invalid index %d", nth))
}

const (
	// hashAggregatorBucketsInitialLen is a guess on how many "items" the
	// 'buckets' map of hashAggregator has the capacity for initially.
	hashAggregatorBucketsInitialLen = 8
	// hashAggregatorSizeOfBucketsItem is a guess on how much space (in bytes)
	// each item added to 'buckets' map of hashAggregator takes up in the map
	// (i.e. it is memory internal to the map, orthogonal to "key-value" pair
	// that we're adding to the map).
	hashAggregatorSizeOfBucketsItem = 64
)

// hashAggregator is a specialization of aggregatorBase that must keep track of
// multiple grouping buckets at a time.
type hashAggregator struct {
	aggregatorBase

	// buckets is used during the accumulation phase to track the bucket keys
	// that have been seen. After accumulation, the keys are extracted into
	// bucketsIter for iteration.
	buckets     map[string]aggregateFuncs
	bucketsIter []string
	// bucketsLenGrowThreshold is the threshold which, when reached by the
	// number of items in 'buckets', will trigger the update to memory
	// accounting. It will start out at hashAggregatorBucketsInitialLen and
	// then will be doubling in size.
	bucketsLenGrowThreshold int
	// alreadyAccountedFor tracks the number of items in 'buckets' memory for
	// which we have already accounted for.
	alreadyAccountedFor int
}

// orderedAggregator is a specialization of aggregatorBase that only needs to
// keep track of a single grouping bucket at a time.
type orderedAggregator struct {
	aggregatorBase

	// bucket is used during the accumulation phase to aggregate results.
	bucket aggregateFuncs
}

var _ execinfra.Processor = &hashAggregator{}
var _ execinfra.RowSource = &hashAggregator{}
var _ execinfra.OpNode = &hashAggregator{}

const hashAggregatorProcName = "hash aggregator"

var _ execinfra.Processor = &orderedAggregator{}
var _ execinfra.RowSource = &orderedAggregator{}
var _ execinfra.OpNode = &orderedAggregator{}

const orderedAggregatorProcName = "ordered aggregator"

// aggregatorState represents the state of the processor.
type aggregatorState int

const (
	aggStateUnknown aggregatorState = iota
	// aggAccumulating means that rows are being read from the input and used to
	// compute intermediary aggregation results.
	aggAccumulating
	// aggEmittingRows means that accumulation has finished and rows are being
	// sent to the output.
	aggEmittingRows
)

func newAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	if spec.IsRowCount() {
		return newCountAggregator(flowCtx, processorID, input, post, output)
	}
	if len(spec.OrderedGroupCols) == len(spec.GroupCols) {
		return newOrderedAggregator(flowCtx, processorID, spec, input, post, output)
	}

	ag := &hashAggregator{
		buckets:                 make(map[string]aggregateFuncs),
		bucketsLenGrowThreshold: hashAggregatorBucketsInitialLen,
	}

	if err := ag.init(
		ag,
		flowCtx,
		processorID,
		spec,
		input,
		post,
		output,
		func(context.Context) []execinfrapb.ProducerMetadata {
			ag.close()
			return nil
		},
	); err != nil {
		return nil, err
	}

	// A new tree.EvalCtx was created during initializing aggregatorBase above
	// and will be used only by this aggregator, so it is ok to update EvalCtx
	// directly.
	ag.EvalCtx.SingleDatumAggMemAccount = &ag.aggFuncsAcc
	return ag, nil
}

func newOrderedAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*orderedAggregator, error) {
	ag := &orderedAggregator{}

	if err := ag.init(
		ag,
		flowCtx,
		processorID,
		spec,
		input,
		post,
		output,
		func(context.Context) []execinfrapb.ProducerMetadata {
			ag.close()
			return nil
		},
	); err != nil {
		return nil, err
	}

	// A new tree.EvalCtx was created during initializing aggregatorBase above
	// and will be used only by this aggregator, so it is ok to update EvalCtx
	// directly.
	ag.EvalCtx.SingleDatumAggMemAccount = &ag.aggFuncsAcc
	return ag, nil
}

// Start is part of the RowSource interface.
func (ag *hashAggregator) Start(ctx context.Context) context.Context {
	return ag.start(ctx, hashAggregatorProcName)
}

// Start is part of the RowSource interface.
func (ag *orderedAggregator) Start(ctx context.Context) context.Context {
	return ag.start(ctx, orderedAggregatorProcName)
}

func (ag *aggregatorBase) start(ctx context.Context, procName string) context.Context {
	ag.input.Start(ctx)
	ctx = ag.StartInternal(ctx, procName)
	ag.cancelChecker = sqlbase.NewCancelChecker(ctx)
	ag.runningState = aggAccumulating
	return ctx
}

func (ag *hashAggregator) close() {
	if ag.InternalClose() {
		log.VEventf(ag.Ctx, 2, "exiting aggregator")
		// If we have started emitting rows, bucketsIter will represent which
		// buckets are still open, since buckets are closed once their results are
		// emitted.
		if ag.bucketsIter == nil {
			for _, bucket := range ag.buckets {
				bucket.close(ag.Ctx)
			}
		} else {
			for _, bucket := range ag.bucketsIter {
				ag.buckets[bucket].close(ag.Ctx)
			}
		}
		// Make sure to release any remaining memory under 'buckets'.
		ag.buckets = nil
		// Note that we should be closing accounts only after closing all the
		// buckets since the latter might be releasing some precisely tracked
		// memory, and if we were to close the accounts first, there would be
		// no memory to release for the buckets.
		ag.bucketsAcc.Close(ag.Ctx)
		ag.aggFuncsAcc.Close(ag.Ctx)
		ag.MemMonitor.Stop(ag.Ctx)
	}
}

func (ag *orderedAggregator) close() {
	if ag.InternalClose() {
		log.VEventf(ag.Ctx, 2, "exiting aggregator")
		if ag.bucket != nil {
			ag.bucket.close(ag.Ctx)
		}
		// Note that we should be closing accounts only after closing the
		// bucket since the latter might be releasing some precisely tracked
		// memory, and if we were to close the accounts first, there would be
		// no memory to release for the bucket.
		ag.bucketsAcc.Close(ag.Ctx)
		ag.aggFuncsAcc.Close(ag.Ctx)
		ag.MemMonitor.Stop(ag.Ctx)
	}
}

// matchLastOrdGroupCols takes a row and matches it with the row stored by
// lastOrdGroupCols. It returns true if the two rows are equal on the grouping
// columns, and false otherwise.
func (ag *aggregatorBase) matchLastOrdGroupCols(row sqlbase.EncDatumRow) (bool, error) {
	for _, colIdx := range ag.orderedGroupCols {
		res, err := ag.lastOrdGroupCols[colIdx].Compare(
			&ag.inputTypes[colIdx], &ag.datumAlloc, ag.EvalCtx, &row[colIdx],
		)
		if res != 0 || err != nil {
			return false, err
		}
	}
	return true, nil
}

// accumulateRows continually reads rows from the input and accumulates them
// into intermediary aggregate results. If it encounters metadata, the metadata
// is immediately returned. Subsequent calls of this function will resume row
// accumulation.
func (ag *hashAggregator) accumulateRows() (
	aggregatorState,
	sqlbase.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	for {
		row, meta := ag.input.Next()
		if meta != nil {
			if meta.Err != nil {
				ag.MoveToDraining(nil /* err */)
				return aggStateUnknown, nil, meta
			}
			return aggAccumulating, nil, meta
		}
		if row == nil {
			log.VEvent(ag.Ctx, 1, "accumulation complete")
			ag.inputDone = true
			break
		}

		if ag.lastOrdGroupCols == nil {
			ag.lastOrdGroupCols = ag.rowAlloc.CopyRow(row)
		} else {
			matched, err := ag.matchLastOrdGroupCols(row)
			if err != nil {
				ag.MoveToDraining(err)
				return aggStateUnknown, nil, nil
			}
			if !matched {
				copy(ag.lastOrdGroupCols, row)
				break
			}
		}
		ag.fpgaAgg.append(row, ag.inputTypes)
		if err := ag.accumulateRow(row); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
	}

	// Queries like `SELECT MAX(n) FROM t` expect a row of NULLs if nothing was
	// aggregated.
	if len(ag.buckets) < 1 && len(ag.groupCols) == 0 {
		bucket, err := ag.createAggregateFuncs()
		if err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
		ag.buckets[""] = bucket
	}

	// Note that, for simplicity, we're ignoring the overhead of the slice of
	// strings.
	if err := ag.bucketsAcc.Grow(ag.Ctx, int64(len(ag.buckets))*sizeOfString); err != nil {
		ag.MoveToDraining(err)
		return aggStateUnknown, nil, nil
	}
	ag.bucketsIter = make([]string, 0, len(ag.buckets))
	for bucket := range ag.buckets {
		ag.bucketsIter = append(ag.bucketsIter, bucket)
	}

	// Transition to aggEmittingRows, and let it generate the next row/meta.
	return aggEmittingRows, nil, nil
}

// accumulateRows continually reads rows from the input and accumulates them
// into intermediary aggregate results. If it encounters metadata, the metadata
// is immediately returned. Subsequent calls of this function will resume row
// accumulation.
func (ag *orderedAggregator) accumulateRows() (
	aggregatorState,
	sqlbase.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	for {
		row, meta := ag.input.Next()
		if meta != nil {
			if meta.Err != nil {
				ag.MoveToDraining(nil /* err */)
				return aggStateUnknown, nil, meta
			}
			return aggAccumulating, nil, meta
		}
		if row == nil {
			log.VEvent(ag.Ctx, 1, "accumulation complete")
			ag.inputDone = true
			break
		}

		if ag.lastOrdGroupCols == nil {
			ag.lastOrdGroupCols = ag.rowAlloc.CopyRow(row)
		} else {
			matched, err := ag.matchLastOrdGroupCols(row)
			if err != nil {
				ag.MoveToDraining(err)
				return aggStateUnknown, nil, nil
			}
			if !matched {
				copy(ag.lastOrdGroupCols, row)
				break
			}
		}
		ag.fpgaAgg.append(row, ag.inputTypes)
		if err := ag.accumulateRow(row); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
	}

	// Queries like `SELECT MAX(n) FROM t` expect a row of NULLs if nothing was
	// aggregated.
	if ag.bucket == nil && ag.isScalar {
		var err error
		ag.bucket, err = ag.createAggregateFuncs()
		if err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
	}

	// Transition to aggEmittingRows, and let it generate the next row/meta.
	return aggEmittingRows, nil, nil
}

// getAggResults returns the new aggregatorState and the results from the
// bucket. The bucket is closed.
func (ag *aggregatorBase) getAggResults(
	bucket aggregateFuncs,
) (aggregatorState, sqlbase.EncDatumRow, *execinfrapb.ProducerMetadata) {
	defer bucket.close(ag.Ctx)
	for i, b := range bucket {
		if b == nil {
			result := ag.fpgaAgg.getResult()
			dr := tree.Datum(tree.NewDInt(tree.DInt(result)))
			ag.row[i] = sqlbase.DatumToEncDatum(&ag.outputTypes[i], dr)
			continue
		}

		result, err := b.Result()
		if err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
		if result == nil {
			// We can't encode nil into an EncDatum, so we represent it with DNull.
			result = tree.DNull
		}
		ag.row[i] = sqlbase.DatumToEncDatum(&ag.outputTypes[i], result)
	}

	if outRow := ag.ProcessRowHelper(ag.row); outRow != nil {
		return aggEmittingRows, outRow, nil
	}
	// We might have switched to draining, we might not have. In case we
	// haven't, aggEmittingRows is accurate. If we have, it will be ignored by
	// the caller.
	return aggEmittingRows, nil, nil
}

// emitRow constructs an output row from an accumulated bucket and returns it.
//
// emitRow() might move to stateDraining. It might also not return a row if the
// ProcOutputHelper filtered the current row out.
func (ag *hashAggregator) emitRow() (
	aggregatorState,
	sqlbase.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	if len(ag.bucketsIter) == 0 {
		// We've exhausted all of the aggregation buckets.
		if ag.inputDone {
			// The input has been fully consumed. Transition to draining so that we
			// emit any metadata that we've produced.
			ag.MoveToDraining(nil /* err */)
			return aggStateUnknown, nil, nil
		}

		// We've only consumed part of the input where the rows are equal over
		// the columns specified by ag.orderedGroupCols, so we need to continue
		// accumulating the remaining rows.

		if err := ag.arena.UnsafeReset(ag.Ctx); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
		// Before we create a new 'buckets' map below, we need to "release" the
		// already accounted for memory of the current map.
		ag.bucketsAcc.Shrink(ag.Ctx, int64(ag.alreadyAccountedFor)*hashAggregatorSizeOfBucketsItem)
		// Note that, for simplicity, we're ignoring the overhead of the slice of
		// strings.
		ag.bucketsAcc.Shrink(ag.Ctx, int64(len(ag.buckets))*sizeOfString)
		ag.bucketsIter = nil
		ag.buckets = make(map[string]aggregateFuncs)
		ag.bucketsLenGrowThreshold = hashAggregatorBucketsInitialLen
		ag.alreadyAccountedFor = 0
		for _, f := range ag.funcs {
			if f.seen != nil {
				f.seen = make(map[string]struct{})
			}
		}

		if err := ag.accumulateRow(ag.lastOrdGroupCols); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}

		return aggAccumulating, nil, nil
	}

	bucket := ag.bucketsIter[0]
	ag.bucketsIter = ag.bucketsIter[1:]

	// Once we get the results from the bucket, we can delete it from the map.
	// This will allow us to return the memory to the system before the hash
	// aggregator is fully done (which matters when we have many buckets).
	// NOTE: accounting for the memory under aggregate builtins in the bucket
	// is updated in getAggResults (the bucket will be closed), however, we
	// choose to not reduce our estimate of the map's internal footprint
	// because it is error-prone to estimate the new footprint (we don't know
	// whether and when Go runtime will release some of the underlying memory).
	// This behavior is ok, though, since actual usage of buckets will be lower
	// than what we accounted for - in the worst case, the query might hit a
	// memory budget limit and error out when it might actually be within the
	// limit. However, we might be under accounting memory usage in other
	// places, so having some over accounting here might be actually beneficial
	// as a defensive mechanism against OOM crashes.
	state, row, meta := ag.getAggResults(ag.buckets[bucket])
	delete(ag.buckets, bucket)
	return state, row, meta
}

// emitRow constructs an output row from an accumulated bucket and returns it.
//
// emitRow() might move to stateDraining. It might also not return a row if the
// ProcOutputHelper filtered a the current row out.
func (ag *orderedAggregator) emitRow() (
	aggregatorState,
	sqlbase.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	if ag.bucket == nil {
		// We've exhausted all of the aggregation buckets.
		if ag.inputDone {
			// The input has been fully consumed. Transition to draining so that we
			// emit any metadata that we've produced.
			ag.MoveToDraining(nil /* err */)
			return aggStateUnknown, nil, nil
		}

		// We've only consumed part of the input where the rows are equal over
		// the columns specified by ag.orderedGroupCols, so we need to continue
		// accumulating the remaining rows.

		if err := ag.arena.UnsafeReset(ag.Ctx); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}
		for _, f := range ag.funcs {
			if f.seen != nil {
				f.seen = make(map[string]struct{})
			}
		}

		if err := ag.accumulateRow(ag.lastOrdGroupCols); err != nil {
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		}

		return aggAccumulating, nil, nil
	}

	bucket := ag.bucket
	ag.bucket = nil
	return ag.getAggResults(bucket)
}

// Next is part of the RowSource interface.
func (ag *hashAggregator) Next() (sqlbase.EncDatumRow, *execinfrapb.ProducerMetadata) {
	for ag.State == execinfra.StateRunning {
		var row sqlbase.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ag.runningState {
		case aggAccumulating:
			ag.runningState, row, meta = ag.accumulateRows()
		case aggEmittingRows:
			ag.runningState, row, meta = ag.emitRow()
		default:
			log.Fatalf(ag.Ctx, "unsupported state: %d", ag.runningState)
		}

		if row == nil && meta == nil {
			continue
		}
		return row, meta
	}
	return nil, ag.DrainHelper()
}

// Next is part of the RowSource interface.
func (ag *orderedAggregator) Next() (sqlbase.EncDatumRow, *execinfrapb.ProducerMetadata) {
	for ag.State == execinfra.StateRunning {
		var row sqlbase.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ag.runningState {
		case aggAccumulating:
			ag.runningState, row, meta = ag.accumulateRows()
		case aggEmittingRows:
			ag.runningState, row, meta = ag.emitRow()
		default:
			log.Fatalf(ag.Ctx, "unsupported state: %d", ag.runningState)
		}

		if row == nil && meta == nil {
			continue
		}
		return row, meta
	}
	return nil, ag.DrainHelper()
}

// ConsumerClosed is part of the RowSource interface.
func (ag *hashAggregator) ConsumerClosed() {
	// The consumer is done, Next() will not be called again.
	ag.close()
}

// ConsumerClosed is part of the RowSource interface.
func (ag *orderedAggregator) ConsumerClosed() {
	// The consumer is done, Next() will not be called again.
	ag.close()
}

func (ag *aggregatorBase) accumulateRowIntoBucket(
	row sqlbase.EncDatumRow, groupKey []byte, bucket aggregateFuncs,
) error {
	var err error
	// Feed the func holders for this bucket the non-grouping datums.
	for i, a := range ag.aggregations {
		if a.FilterColIdx != nil {
			col := *a.FilterColIdx
			if err := row[col].EnsureDecoded(&ag.inputTypes[col], &ag.datumAlloc); err != nil {
				return err
			}
			if row[*a.FilterColIdx].Datum != tree.DBoolTrue {
				// This row doesn't contribute to this aggregation.
				continue
			}
		}
		// Extract the corresponding arguments from the row to feed into the
		// aggregate function.
		// Most functions require at most one argument thus we separate
		// the first argument and allocation of (if applicable) a variadic
		// collection of arguments thereafter.
		var firstArg tree.Datum
		var otherArgs tree.Datums
		if len(a.ColIdx) > 1 {
			otherArgs = make(tree.Datums, len(a.ColIdx)-1)
		}
		isFirstArg := true
		for j, c := range a.ColIdx {
			if err := row[c].EnsureDecoded(&ag.inputTypes[c], &ag.datumAlloc); err != nil {
				return err
			}
			if isFirstArg {
				firstArg = row[c].Datum
				isFirstArg = false
				continue
			}
			otherArgs[j-1] = row[c].Datum
		}

		canAdd := true
		if a.Distinct {
			canAdd, err = ag.funcs[i].isDistinct(
				ag.Ctx,
				&ag.datumAlloc,
				groupKey,
				firstArg,
				otherArgs,
			)
			if err != nil {
				return err
			}
		}
		if !canAdd {
			continue
		}
		if err := bucket[i].Add(ag.Ctx, firstArg, otherArgs...); err != nil {
			return err
		}
	}
	return nil
}

// accumulateRow accumulates a single row, returning an error if accumulation
// failed for any reason.
func (ag *hashAggregator) accumulateRow(row sqlbase.EncDatumRow) error {
	if err := ag.cancelChecker.Check(); err != nil {
		return err
	}

	// The encoding computed here determines which bucket the non-grouping
	// datums are accumulated to.
	encoded, err := ag.encode(ag.scratch, row)
	if err != nil {
		return err
	}
	ag.scratch = encoded[:0]

	bucket, ok := ag.buckets[string(encoded)]
	if !ok {
		s, err := ag.arena.AllocBytes(ag.Ctx, encoded)
		if err != nil {
			return err
		}
		bucket, err = ag.createAggregateFuncs()
		if err != nil {
			return err
		}
		ag.buckets[s] = bucket
		if len(ag.buckets) == ag.bucketsLenGrowThreshold {
			toAccountFor := ag.bucketsLenGrowThreshold - ag.alreadyAccountedFor
			if err := ag.bucketsAcc.Grow(ag.Ctx, int64(toAccountFor)*hashAggregatorSizeOfBucketsItem); err != nil {
				return err
			}
			ag.alreadyAccountedFor = ag.bucketsLenGrowThreshold
			ag.bucketsLenGrowThreshold *= 2
		}
	}

	return ag.accumulateRowIntoBucket(row, encoded, bucket)
}

// accumulateRow accumulates a single row, returning an error if accumulation
// failed for any reason.
func (ag *orderedAggregator) accumulateRow(row sqlbase.EncDatumRow) error {
	if err := ag.cancelChecker.Check(); err != nil {
		return err
	}

	if ag.bucket == nil {
		var err error
		ag.bucket, err = ag.createAggregateFuncs()
		if err != nil {
			return err
		}
	}

	return ag.accumulateRowIntoBucket(row, nil /* groupKey */, ag.bucket)
}

type aggregateFuncHolder struct {
	create func(*tree.EvalContext, tree.Datums) tree.AggregateFunc

	// arguments is the list of constant (non-aggregated) arguments to the
	// aggregate, for instance, the separator in string_agg.
	arguments tree.Datums

	group *aggregatorBase
	seen  map[string]struct{}
	arena *stringarena.Arena
}

const (
	sizeOfString         = int64(unsafe.Sizeof(""))
	sizeOfAggregateFuncs = int64(unsafe.Sizeof(aggregateFuncs{}))
	sizeOfAggregateFunc  = int64(unsafe.Sizeof(tree.AggregateFunc(nil)))
)

func (ag *aggregatorBase) newAggregateFuncHolder(
	create func(*tree.EvalContext, tree.Datums) tree.AggregateFunc, arguments tree.Datums,
) *aggregateFuncHolder {
	return &aggregateFuncHolder{
		create:    create,
		group:     ag,
		arena:     &ag.arena,
		arguments: arguments,
	}
}

// isDistinct returns whether this aggregateFuncHolder has not already seen the
// encoding of grouping columns and argument columns. It should be used *only*
// when we have DISTINCT aggregation so that we can aggregate only the "first"
// row in the group.
func (a *aggregateFuncHolder) isDistinct(
	ctx context.Context,
	alloc *sqlbase.DatumAlloc,
	prefix []byte,
	firstArg tree.Datum,
	otherArgs tree.Datums,
) (bool, error) {
	// Allocate one EncDatum that will be reused when encoding every argument.
	ed := sqlbase.EncDatum{Datum: firstArg}
	encoded, err := ed.Fingerprint(firstArg.ResolvedType(), alloc, prefix)
	if err != nil {
		return false, err
	}
	if otherArgs != nil {
		for _, arg := range otherArgs {
			ed.Datum = arg
			encoded, err = ed.Fingerprint(arg.ResolvedType(), alloc, encoded)
			if err != nil {
				return false, err
			}
		}
	}

	if _, ok := a.seen[string(encoded)]; ok {
		// We have already seen a row with such combination of grouping and
		// argument columns.
		return false, nil
	}
	s, err := a.arena.AllocBytes(ctx, encoded)
	if err != nil {
		return false, err
	}
	a.seen[s] = struct{}{}
	return true, nil
}

// encode returns the encoding for the grouping columns, this is then used as
// our group key to determine which bucket to add to.
func (ag *aggregatorBase) encode(
	appendTo []byte, row sqlbase.EncDatumRow,
) (encoding []byte, err error) {
	for _, colIdx := range ag.groupCols {
		appendTo, err = row[colIdx].Fingerprint(
			&ag.inputTypes[colIdx], &ag.datumAlloc, appendTo)
		if err != nil {
			return appendTo, err
		}
	}
	return appendTo, nil
}

func (ag *aggregatorBase) createAggregateFuncs() (aggregateFuncs, error) {
	if err := ag.bucketsAcc.Grow(ag.Ctx, sizeOfAggregateFuncs+sizeOfAggregateFunc*int64(len(ag.funcs))); err != nil {
		return nil, err
	}
	bucket := make(aggregateFuncs, len(ag.funcs))
	for i, f := range ag.funcs {
		if f == nil {
			bucket[i] = nil
			continue
		}
		agg := f.create(ag.EvalCtx, f.arguments)
		if err := ag.bucketsAcc.Grow(ag.Ctx, agg.Size()); err != nil {
			return nil, err
		}
		bucket[i] = agg
	}
	return bucket, nil
}
