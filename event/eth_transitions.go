package event

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/Oneledger/protocol/data/ethereum"
	"github.com/Oneledger/protocol/utils/transition"
)

func init() {
	EthEngine = transition.NewEngine(
		[]transition.Status{
			transition.Status(ethereum.New),
			transition.Status(ethereum.BusyBroadcasting),
			transition.Status(ethereum.BusyFinalizing),
			transition.Status(ethereum.Finalized),
			transition.Status(ethereum.Minted),
		})

	err := EthEngine.Register(transition.Transition{
		Name: ethereum.BROADCASTING,
		Fn:   Broadcasting,
		From: transition.Status(ethereum.New),
		To:   transition.Status(ethereum.BusyBroadcasting),
	})
	if err != nil {
		panic(err)
	}

	err = EthEngine.Register(transition.Transition{
		Name: ethereum.FINALIZING,
		Fn:   Finalizing,
		From: transition.Status(ethereum.BusyBroadcasting),
		To:   transition.Status(ethereum.BusyFinalizing),
	})
	if err != nil {
		panic(err)
	}

	err = EthEngine.Register(transition.Transition{
		Name: ethereum.FINALIZE,
		Fn:   Finalization,
		From: transition.Status(ethereum.BusyFinalizing),
		To:   transition.Status(ethereum.Finalized),
	})
	if err != nil {
		panic(err)
	}

	err = EthEngine.Register(transition.Transition{
		Name: ethereum.MINTING,
		Fn:   Minting,
		From: transition.Status(ethereum.Finalized),
		To:   transition.Status(ethereum.Minted),
	})
	if err != nil {
		panic(err)
	}
	err = EthEngine.Register(transition.Transition{
		Name: ethereum.CLEANUP,
		Fn:   Cleanup,
		From: transition.Status(ethereum.Minted),
		To:   0,
	})
	if err != nil {
		panic(err)
	}
}

//TODO Go back to Busy broadcasting if there is a failure in Finalizing.
func Broadcasting(ctx interface{}) error {
	context, ok := ctx.(ethereum.TrackerCtx)
	if !ok {
		return errors.New("error casting tracker context")
	}

	fmt.Println("broadcasting")
	tracker := context.Tracker

	if tracker.State != ethereum.New {
		err := errors.New("Cannot Broadcast from the current state")
		return errors.Wrap(err, string(tracker.State))
	}

	tracker.State = ethereum.BusyBroadcasting

	//create broadcasting job
	job := NewETHBroadcast(tracker.TrackerName, tracker.State)
	err := context.JobStore.SaveJob(job)
	if err != nil {
		return errors.Wrap(errors.New("job serialization failed err: "), err.Error())
	}

	return nil
}

func Finalizing(ctx interface{}) error {
	context, ok := ctx.(ethereum.TrackerCtx)
	if !ok {
		return errors.New("error casting tracker context")
	}

	fmt.Println("finalizing")

	tracker := context.Tracker

	if tracker.State != ethereum.BusyBroadcasting {
		err := errors.New("Cannot start Finalizing from the current state")
		return errors.Wrap(err, string(tracker.State))
	}

	//Check if current Node voted
	voted := tracker.CheckIfVoted(context.CurrNodeAddr)

	if !voted {
		//Check Broadcasting job
		bjob, err := context.JobStore.GetJob(tracker.GetJobID(ethereum.BusyBroadcasting))
		if err != nil {
			return errors.Wrap(err, "failed to get job")
		}

		if bjob.IsDone() {

			//Create job to check finality
			job := NewETHCheckFinality(tracker.TrackerName, ethereum.BusyFinalizing)
			err := context.JobStore.SaveJob(job)
			if err != nil {
				return errors.Wrap(errors.New("job serialization failed err: "), err.Error())
			}
		}
	}

	numVotes := tracker.GetVotes()

	if numVotes > 0 {
		tracker.State = ethereum.BusyFinalizing
	}

	return nil
}

func Finalization(ctx interface{}) error {
	context, ok := ctx.(ethereum.TrackerCtx)
	if !ok {
		return errors.New("error casting tracker context")
	}

	tracker := context.Tracker

	if tracker.State != ethereum.BusyFinalizing {
		err := errors.New("cannot finalize from the current state")
		return errors.Wrap(err, string(tracker.State))
	}

	//Check if current Node voted
	voted := tracker.CheckIfVoted(context.CurrNodeAddr)

	if !voted {
		//Create job to check finality
		job := NewETHCheckFinality(tracker.TrackerName, tracker.State)
		err := context.JobStore.SaveJob(job)
		if err != nil {
			return errors.Wrap(errors.New("job serialization failed err: "), err.Error())
		}
	}

	if tracker.Finalized() {
		tracker.State = ethereum.Finalized
	}
	return nil
}

func Minting(ctx interface{}) error {
	context, ok := ctx.(ethereum.TrackerCtx)
	if !ok {
		return errors.New("error casting tracker context")
	}

	tracker := context.Tracker

	if tracker.State != ethereum.Finalized {
		err := errors.New("Cannot Mint from the current state")
		return errors.Wrap(err, string(tracker.State))
	}
	//todo: create a job to mint

	if tracker.TaskCompleted {
		tracker.State = ethereum.Minted
	}
	return nil
}

func Cleanup(ctx interface{}) error {
	context, ok := ctx.(ethereum.TrackerCtx)
	if !ok {
		return errors.New("error casting tracker context")
	}

	tracker := context.Tracker
	//todo: delete the tracker and jobs related

	//Delete Broadcasting Job
	bjob, err := context.JobStore.GetJob(tracker.GetJobID(ethereum.BusyBroadcasting))
	if err != nil {
		return errors.Wrap(err, "failed to get job")
	}

	err = context.JobStore.DeleteJob(bjob)
	if err != nil {
		return err
	}

	//Delete CheckFinality Job
	fjob, err := context.JobStore.GetJob(tracker.GetJobID(ethereum.BusyBroadcasting))
	if err != nil {
		return errors.Wrap(err, "failed to get job")
	}
	err = context.JobStore.DeleteJob(fjob)
	if err != nil {
		return err
	}

	//Delete Tracker
	res, err := context.TrackerStore.Delete(tracker.TrackerName)
	if err != nil || !res {
		return err
	}

	return nil
}
