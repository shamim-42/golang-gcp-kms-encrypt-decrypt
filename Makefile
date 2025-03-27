# Define the environment-specific Makefiles
DEV_MAKEFILE := Makefile.dev

dev_postgres:
	@make -f $(DEV_MAKEFILE) postgres

dev_createdb:
	@make -f $(DEV_MAKEFILE) createdb

dev_dropdb:
	@make -f $(DEV_MAKEFILE) dropdb

dev_reset_db:
	@make -f $(DEV_MAKEFILE) reset_db

dev_delete_volume:
	@make -f $(DEV_MAKEFILE) delete_volume

dev_air:
	@make -f $(DEV_MAKEFILE) air

.PHONY: dev_postgres dev_createdb dev_dropdb dev_reset_db dev_delete_volume dev_air 