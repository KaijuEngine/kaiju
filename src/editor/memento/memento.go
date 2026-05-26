/******************************************************************************/
/* memento.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package memento

type Memento interface {
	Redo()   // Called to redo the action
	Undo()   // Called to undo the action
	Delete() // Called when the undo state is overridden by new undo data
	Exit()   // Called when the undo state falls off the history list
}
