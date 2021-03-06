package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ovh/cds/engine/api/application"
	"github.com/ovh/cds/engine/api/project"
	"github.com/ovh/cds/engine/api/workflow"
	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/log"
)

func (api *API) getApplicationDeploymentStrategiesConfigHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars["key"]
		appName := vars["permApplicationName"]

		app, err := application.LoadByName(api.mustDB(), api.Cache, key, appName, getUser(ctx), application.LoadOptions.WithDeploymentStrategies)
		if err != nil {
			return sdk.WrapError(err, "getApplicationDeploymentStrategiesConfigHandler")
		}

		return service.WriteJSON(w, app.DeploymentStrategies, http.StatusOK)
	}
}

func (api *API) postApplicationDeploymentStrategyConfigHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars["key"]
		appName := vars["permApplicationName"]
		pfName := vars["platform"]

		var pfConfig sdk.PlatformConfig
		if err := service.UnmarshalBody(r, &pfConfig); err != nil {
			return err
		}

		tx, errtx := api.mustDB().Begin()
		if errtx != nil {
			return errtx
		}
		defer tx.Rollback()

		proj, err := project.Load(tx, api.Cache, key, getUser(ctx), project.LoadOptions.WithPlatforms)
		if err != nil {
			return sdk.WrapError(err, "unable to load project")
		}

		var pf *sdk.ProjectPlatform
		for i := range proj.Platforms {
			if proj.Platforms[i].Name == pfName {
				pf = &proj.Platforms[i]
				break
			}
		}

		if pf == nil {
			return sdk.WrapError(sdk.ErrNotFound, "postApplicationDeploymentStrategyConfigHandler> platform not found on project")
		}

		if !pf.Model.Deployment {
			return sdk.WrapError(sdk.ErrForbidden, "postApplicationDeploymentStrategyConfigHandler> platform doesn't support deployment")
		}

		app, err := application.LoadByName(tx, api.Cache, key, appName, getUser(ctx), application.LoadOptions.WithClearDeploymentStrategies)
		if err != nil {
			return sdk.WrapError(err, "unable to load application")
		}

		oldPfConfig, has := app.DeploymentStrategies[pfName]
		if !has {
			oldPfConfig = pf.Model.DeploymentDefaultConfig
		}
		if oldPfConfig == nil {
			oldPfConfig = sdk.PlatformConfig{}
		}
		oldPfConfig.MergeWith(pfConfig)
		pfConfig = oldPfConfig

		if err := application.SetDeploymentStrategy(tx, proj.ID, app.ID, pf.Model.ID, pfName, pfConfig); err != nil {
			return sdk.WrapError(err, "postApplicationDeploymentStrategyConfigHandler")
		}

		app, err = application.LoadByName(tx, api.Cache, key, appName, getUser(ctx), application.LoadOptions.WithDeploymentStrategies)
		if err != nil {
			return sdk.WrapError(err, "unable to load application")
		}

		if err := tx.Commit(); err != nil {
			return sdk.WrapError(err, "unable to commit tx")
		}

		if getProvider(ctx) != nil {
			p := getProvider(ctx)
			log.Info("postApplicationDeploymentStrategyConfigHandler> application %s configuration successfully updated by provider %s", appName, *p)
		}

		return service.WriteJSON(w, app, http.StatusOK)
	}
}

func (api *API) deleteApplicationDeploymentStrategyConfigHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars["key"]
		appName := vars["permApplicationName"]
		pfName := vars["platform"]

		tx, errtx := api.mustDB().Begin()
		if errtx != nil {
			return errtx
		}
		defer tx.Rollback()

		proj, err := project.Load(tx, api.Cache, key, getUser(ctx), project.LoadOptions.WithPlatforms)
		if err != nil {
			return sdk.WrapError(err, "unable to load project")
		}

		var pf *sdk.ProjectPlatform
		for i := range proj.Platforms {
			if proj.Platforms[i].Name == pfName {
				pf = &proj.Platforms[i]
				break
			}
		}

		if pf == nil {
			return sdk.WrapError(sdk.ErrNotFound, "deleteApplicationDeploymentStrategyConfigHandler> platform not found on project")
		}

		if !pf.Model.Deployment {
			return sdk.WrapError(sdk.ErrForbidden, "deleteApplicationDeploymentStrategyConfigHandler> platform doesn't support deployment")
		}

		app, err := application.LoadByName(tx, api.Cache, key, appName, getUser(ctx), application.LoadOptions.WithDeploymentStrategies)
		if err != nil {
			return sdk.WrapError(err, "unable to load application")
		}

		isUsed, err := workflow.IsDeploymentPlatformUsed(tx, proj.ID, app.ID, pfName)
		if err != nil {
			return sdk.WrapError(err, "unable to check if platform is used")
		}

		if isUsed {
			return sdk.NewError(sdk.ErrForbidden, fmt.Errorf("platform is still used in a workflow"))
		}

		if _, has := app.DeploymentStrategies[pfName]; !has {
			return sdk.WrapError(sdk.ErrNotFound, "deleteApplicationDeploymentStrategyConfigHandler> unable to find strategy")
		}

		delete(app.DeploymentStrategies, pfName)
		if err := application.DeleteDeploymentStrategy(tx, proj.ID, app.ID, pf.ID); err != nil {
			return sdk.WrapError(err, "deleteApplicationDeploymentStrategyConfigHandler")
		}

		if err := tx.Commit(); err != nil {
			return sdk.WrapError(err, "unable to commit tx")
		}

		return service.WriteJSON(w, app, http.StatusOK)
	}
}

func (api *API) getApplicationDeploymentStrategyConfigHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars["key"]
		appName := vars["permApplicationName"]
		pfName := vars["platform"]
		withClearPassword := FormBool(r, "withClearPassword")

		opts := []application.LoadOptionFunc{
			application.LoadOptions.WithDeploymentStrategies,
		}
		if withClearPassword {
			opts = []application.LoadOptionFunc{application.LoadOptions.WithClearDeploymentStrategies}
		}

		app, err := application.LoadByName(api.mustDB(), api.Cache, key, appName, getUser(ctx), opts...)
		if err != nil {
			return sdk.WrapError(err, "getApplicationDeploymentStrategyConfigHandler")
		}

		cfg, ok := app.DeploymentStrategies[pfName]
		if !ok {
			return sdk.ErrNotFound
		}

		return service.WriteJSON(w, cfg, http.StatusOK)
	}
}
