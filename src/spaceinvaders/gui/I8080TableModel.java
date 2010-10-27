/*
 *  Copyright 2010 Barend Scholtus
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package spaceinvaders.gui;

import javax.swing.table.AbstractTableModel;
import spaceinvaders.I8080Context;
import spaceinvaders.I8080OpInfo;

/**
 *
 * @author Barend Scholtus
 */
public class I8080TableModel extends AbstractTableModel {

    private int[] memory;
    private boolean[] isRowAnOpCode;

    private String[] columnNames = new String[] {
        "Addr", "Data", "Opcode", "Operand", "Description"
    };

    private Class[] columnClass = new Class[] {
        Integer.class, Integer.class, String.class, String.class, String.class
    };

    public I8080TableModel(I8080Context ctx) {
        this.memory = ctx.getMemory();
        isRowAnOpCode = new boolean[memory.length];
    }

    public int getColumnCount() {
        return columnNames.length;
    }

    public int getRowCount() {
        return memory.length;
    }

    public Object getValueAt(int rowIndex, int columnIndex) {
        switch (columnIndex) {
            case 0:
                return rowIndex;
            case 1:
                return memory[rowIndex];
            case 2:
                return isRowAnOpCode[rowIndex]
                        ? I8080OpInfo.opCode[memory[rowIndex]].mnemonic
                        : "";
            case 3:
                if (!isRowAnOpCode[rowIndex]) {
                    return "";
                }
                I8080OpInfo opInfo = I8080OpInfo.opCode[memory[rowIndex]];
                switch (opInfo.length) {
                    case 2:
                        return opInfo.operands + " ("
                                + Util.toHexString(memory[rowIndex+1], 2) + ")";
                    case 3:
                        return opInfo.operands + " ("
                                + Util.toHexString(memory[rowIndex+2], 2)
                                + Util.toHexString(memory[rowIndex+1], 2) + ")";
                    default:
                        return opInfo.operands;
                }
            case 4:
                return isRowAnOpCode[rowIndex]
                        ? I8080OpInfo.opCode[memory[rowIndex]].description
                        : "";
        }
        return "";
    }

    @Override
    public String getColumnName(int col) {
        return columnNames[col];
    }

    @Override
    public Class getColumnClass(int c) {
        return columnClass[c];
    }

    @Override
    public boolean isCellEditable(int row, int col) {
        return false;
    }

    public void setIsRowAnOpCode(int row) {
        isRowAnOpCode[row] = true;
        fireTableRowsUpdated(row, row);
    }
}
