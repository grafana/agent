package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// Action represents an action in the Hetzner Cloud.
type Action struct {
	ID           int
	Status       ActionStatus
	Command      string
	Progress     int
	Started      time.Time
	Finished     time.Time
	ErrorCode    string
	ErrorMessage string
	Resources    []*ActionResource
}

// ActionStatus represents an action's status.
type ActionStatus string

// List of action statuses.
const (
	ActionStatusRunning ActionStatus = "running"
	ActionStatusSuccess ActionStatus = "success"
	ActionStatusError   ActionStatus = "error"
)

// ActionResource references other resources from an action.
type ActionResource struct {
	ID   int
	Type ActionResourceType
}

// ActionResourceType represents an action's resource reference type.
type ActionResourceType string

// List of action resource reference types.
const (
	ActionResourceTypeServer     ActionResourceType = "server"
	ActionResourceTypeImage      ActionResourceType = "image"
	ActionResourceTypeISO        ActionResourceType = "iso"
	ActionResourceTypeFloatingIP ActionResourceType = "floating_ip"
	ActionResourceTypeVolume     ActionResourceType = "volume"
)

// ActionError is the error of an action.
type ActionError struct {
	Code    string
	Message string
}

func (e ActionError) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func (a *Action) Error() error {
	if a.ErrorCode != "" && a.ErrorMessage != "" {
		return ActionError{
			Code:    a.ErrorCode,
			Message: a.ErrorMessage,
		}
	}
	return nil
}

// ActionClient is a client for the actions API.
type ActionClient struct {
	client *Client
}

// GetByID retrieves an action by its ID. If the action does not exist, nil is returned.
func (c *ActionClient) GetByID(ctx context.Context, id int) (*Action, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/actions/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ActionGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return ActionFromSchema(body.Action), resp, nil
}

// ActionListOpts specifies options for listing actions.
type ActionListOpts struct {
	ListOpts
	ID     []int
	Status []ActionStatus
	Sort   []string
}

func (l ActionListOpts) values() url.Values {
	vals := l.ListOpts.values()
	for _, id := range l.ID {
		vals.Add("id", fmt.Sprintf("%d", id))
	}
	for _, status := range l.Status {
		vals.Add("status", string(status))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of actions for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ActionClient) List(ctx context.Context, opts ActionListOpts) ([]*Action, *Response, error) {
	path := "/actions?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ActionListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	actions := make([]*Action, 0, len(body.Actions))
	for _, i := range body.Actions {
		actions = append(actions, ActionFromSchema(i))
	}
	return actions, resp, nil
}

// All returns all actions.
func (c *ActionClient) All(ctx context.Context) ([]*Action, error) {
	allActions := []*Action{}

	opts := ActionListOpts{}
	opts.PerPage = 50

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		actions, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allActions = append(allActions, actions...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allActions, nil
}

// AllWithOpts returns all actions for the given options.
func (c *ActionClient) AllWithOpts(ctx context.Context, opts ActionListOpts) ([]*Action, error) {
	allActions := []*Action{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		actions, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allActions = append(allActions, actions...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allActions, nil
}

// WatchOverallProgress watches several actions' progress until they complete with success or error.
func (c *ActionClient) WatchOverallProgress(ctx context.Context, actions []*Action) (<-chan int, <-chan error) {
	errCh := make(chan error, len(actions))
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		successIDs := make([]int, 0, len(actions))
		watchIDs := make(map[int]struct{}, len(actions))
		for _, action := range actions {
			watchIDs[action.ID] = struct{}{}
		}

		ticker := time.NewTicker(c.client.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-ticker.C:
				break
			}

			opts := ActionListOpts{}
			for watchID := range watchIDs {
				opts.ID = append(opts.ID, watchID)
			}

			as, err := c.AllWithOpts(ctx, opts)
			if err != nil {
				errCh <- err
				return
			}

			for _, a := range as {
				switch a.Status {
				case ActionStatusRunning:
					continue
				case ActionStatusSuccess:
					delete(watchIDs, a.ID)
					successIDs := append(successIDs, a.ID)
					sendProgress(progressCh, int(float64(len(actions)-len(successIDs))/float64(len(actions))*100))
				case ActionStatusError:
					delete(watchIDs, a.ID)
					errCh <- fmt.Errorf("action %d failed: %w", a.ID, a.Error())
				}
			}

			if len(watchIDs) == 0 {
				return
			}
		}
	}()

	return progressCh, errCh
}

// WatchProgress watches one action's progress until it completes with success or error.
func (c *ActionClient) WatchProgress(ctx context.Context, action *Action) (<-chan int, <-chan error) {
	errCh := make(chan error, 1)
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		ticker := time.NewTicker(c.client.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-ticker.C:
				break
			}

			a, _, err := c.GetByID(ctx, action.ID)
			if err != nil {
				errCh <- err
				return
			}

			switch a.Status {
			case ActionStatusRunning:
				sendProgress(progressCh, a.Progress)
			case ActionStatusSuccess:
				sendProgress(progressCh, 100)
				errCh <- nil
				return
			case ActionStatusError:
				errCh <- a.Error()
				return
			}
		}
	}()

	return progressCh, errCh
}

func sendProgress(progressCh chan int, p int) {
	select {
	case progressCh <- p:
		break
	default:
		break
	}
}
