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

import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import javax.swing.*;
import javax.swing.table.*;
import spaceinvaders.*;


/**
 *
 * @author Barend Scholtus
 */
public class I8080Panel extends JPanel implements ActionListener {

    private JLabel lblPC = new JLabel("PC=");
    private JLabel lblSP = new JLabel("SP=");
    private JLabel lblAF = new JLabel("AF=");
    private JLabel lblBC = new JLabel("BC=");
    private JLabel lblDE = new JLabel("DE=");
    private JLabel lblHL = new JLabel("HL=");

    private JTextField fldPC = new JTextField(4);
    private JTextField fldSP = new JTextField(4);
    private JTextField fldAF = new JTextField(4);
    private JTextField fldBC = new JTextField(4);
    private JTextField fldDE = new JTextField(4);
    private JTextField fldHL = new JTextField(4);

    private JLabel lblFlags = new JLabel("SZAPC=");
    private JTextField fldFlags = new JTextField(4);

    private int cycles = 0;
    private JLabel lblCycles = new JLabel("cycles=");
    private JTextField fldCycles = new JTextField(8);

    private JButton btnStep = new JButton("Step");
    private JButton btnRunTC = new JButton("Run to Cursor");
    private JButton btnStop = new JButton("Stop");
    private JButton btnReset = new JButton("Reset");

    private JScrollPane scrollPane;
    private JTable table;
    private I8080TableModel tableModel;

    private volatile Thread runToCursorThread;

    private I8080Context ctx;

    public I8080Panel(I8080Context ctx) {
        this.ctx = ctx;
        createComponents();
        btnStep.addActionListener(this);
        btnRunTC.addActionListener(this);
        btnStop.addActionListener(this);
        btnReset.addActionListener(this);
        updateDisplay();
    }

    private void createComponents() {
        fldPC.setEditable(false);
        fldSP.setEditable(false);
        fldAF.setEditable(false);
        fldBC.setEditable(false);
        fldDE.setEditable(false);
        fldHL.setEditable(false);
        fldFlags.setEditable(false);
        fldCycles.setEditable(false);

        JPanel firstPanel = new JPanel(new FlowLayout(FlowLayout.LEFT));
        JPanel secondPanel = new JPanel(new FlowLayout(FlowLayout.LEFT));
        JPanel thirdPanel = new JPanel(new FlowLayout(FlowLayout.LEFT));

        firstPanel.add(lblPC);
        firstPanel.add(fldPC);
        firstPanel.add(lblSP);
        firstPanel.add(fldSP);
        firstPanel.add(lblAF);
        firstPanel.add(fldAF);
        firstPanel.add(lblFlags);
        firstPanel.add(fldFlags);

        secondPanel.add(lblBC);
        secondPanel.add(fldBC);
        secondPanel.add(lblDE);
        secondPanel.add(fldDE);
        secondPanel.add(lblHL);
        secondPanel.add(fldHL);

        thirdPanel.add(lblCycles);
        thirdPanel.add(fldCycles);
        thirdPanel.add(btnStep);
        thirdPanel.add(btnRunTC);
        thirdPanel.add(btnStop);
        thirdPanel.add(btnReset);

        JPanel topPanel = new JPanel(new GridLayout(3, 1, 0, 0));
        topPanel.add(firstPanel);
        topPanel.add(secondPanel);
        topPanel.add(thirdPanel);

        tableModel = new I8080TableModel(ctx);
        table = new JTable(tableModel);
        scrollPane = new JScrollPane(table);

        table.getColumnModel().getColumn(0)
                .setCellRenderer(new Hex4Renderer());
        table.getColumnModel().getColumn(1)
                .setCellRenderer(new Hex2Renderer());
        tableModel.setInitialColumnWidths(table);
        table.setColumnSelectionAllowed(false);
        table.setRowSelectionAllowed(true);
        table.setSelectionMode(ListSelectionModel.SINGLE_SELECTION);
        table.setShowHorizontalLines(false);
        table.setAutoResizeMode(JTable.AUTO_RESIZE_OFF);

        setLayout(new BorderLayout());
        add(topPanel, BorderLayout.NORTH);
        add(scrollPane);
    }

    private void updateDisplay() {
        I8080 cpu = ctx.getCpu();
        
        fldPC.setText(Hex4Renderer.toHexString(cpu.PC()));
        fldSP.setText(Hex4Renderer.toHexString(cpu.SP()));
        fldAF.setText(Hex4Renderer.toHexString(cpu.AF()));
        fldBC.setText(Hex4Renderer.toHexString(cpu.BC()));
        fldDE.setText(Hex4Renderer.toHexString(cpu.DE()));
        fldHL.setText(Hex4Renderer.toHexString(cpu.HL()));
        fldFlags.setText((cpu.isCarry()?"1":"0")
                + (cpu.isParity()?"1":"0") + (cpu.isAuxCarry()?"1":"0")
                + (cpu.isZero()?"1":"0") + (cpu.isSign()?"1":"0"));
        fldCycles.setText(Integer.toString(cycles));

        ListSelectionModel lsm = table.getSelectionModel();
        lsm.setSelectionInterval(cpu.PC(), cpu.PC());
        table.scrollRectToVisible(table.getCellRect(cpu.PC(), 0, true));
    }

    public static String toBinaryString8(int val) {
        return (val<=0x01 ? "0000000" :
                val<=0x02 ? "000000" :
                val<=0x04 ? "00000" :
                val<=0x08 ? "0000" :
                val<=0x10 ? "000" :
                val<=0x20 ? "00" :
                val<=0x40 ? "0" : "")
                    + Integer.toBinaryString(val);
    }

    private void reset() {
        ctx.getCpu().reset();
        cycles = 0;
        updateDisplay();
    }

    public void actionPerformed(ActionEvent e) {
        Object src = e.getSource();

        if (src == btnStep) {
            I8080 cpu = ctx.getCpu();
            cycles += cpu.instruction();

            // if next op has params, disable param rows in GUI
            int nextOp = ctx.getMemory()[cpu.PC()];
            int opType = I8080OpInfo.opInfo[nextOp].getOperandType();
            if (opType == I8080OpInfo.OPND_R_DATA
                    || opType == I8080OpInfo.OPND_M_DATA) {
                tableModel.setIsRowData(cpu.PC()+1, 1);
            }
            if (opType == I8080OpInfo.OPND_RP_DATA16
                    || opType == I8080OpInfo.OPND_ADDR) {
                tableModel.setIsRowData(cpu.PC()+1, 2);
            }

            updateDisplay();
        } else if (src == btnRunTC) {
            int row = table.getSelectedRow();
            startRunToCursor(row);
        } else if (src == btnStop) {
            stopRunToCursor();
        } else if (src == btnReset) {
            reset();
            updateDisplay();
        }
    }
    
    private void startRunToCursor(final int toPC) {
        runToCursorThread = new Thread() {
            @Override
            public void run() {
                Thread curr = Thread.currentThread();
                while (runToCursorThread == curr && toPC != ctx.getCpu().PC()) {
                    cycles += ctx.getCpu().instruction();
                    updateDisplay();
                }
            }
        };
        runToCursorThread.start();
    }
    
    private void stopRunToCursor() {
        runToCursorThread = null;
    }
}

class I8080TableModel extends AbstractTableModel {

    private int[] memory;
    private boolean[] isRowData;

    private String[] columnNames = new String[] {
        "Addr", "Data", "Opcode", "Operand", "Description"
    };

    private Class[] columnClass = new Class[] {
        Integer.class, Integer.class, String.class, String.class, String.class
    };

    private int[] initialColumnsWidth = new int[] {
        40, 30, 45, 60
    };

    public I8080TableModel(I8080Context ctx) {
        this.memory = ctx.getMemory();
        isRowData = new boolean[memory.length];
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
                return isRowData[rowIndex] ? ""
                        : I8080OpInfo.opInfo[memory[rowIndex]].getMnemonic();
            case 3:
                return isRowData[rowIndex] ? ""
                        : I8080OpInfo.getOperandsAsString(memory, rowIndex);
            case 4:
                return isRowData[rowIndex] ? ""
                        : I8080OpInfo.opInfo[memory[rowIndex]].getDescription();
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

    public void setInitialColumnWidths(JTable table) {
        int tot = 0;
        for (int i = 0; i < table.getColumnCount()-1; i++) {
            table.getColumnModel().getColumn(i).setPreferredWidth(initialColumnsWidth[i]);
            tot += initialColumnsWidth[i];
        }
        table.getColumnModel().getColumn(table.getColumnCount()-1)
                .setPreferredWidth(table.getPreferredScrollableViewportSize().width-tot);
    }

    public void setIsRowData(int row, int count) {
        for (int i=0; i<count; i++) {
            isRowData[row++] = true;
        }
        fireTableRowsUpdated(row, row+count-1);
    }
}

class Hex2Renderer extends DefaultTableCellRenderer {

    @Override
    public void setValue(Object value) {
        if (!(value instanceof Integer) || (value == null)) {
            setText("##");
        }
        setText(toHexString((Integer)value));
    }

    public static String toHexString(int i) {
        return (i < 0x10 ? "0" : "")
                    + Integer.toHexString(i).toUpperCase();
    }
}

class Hex4Renderer extends DefaultTableCellRenderer {

    @Override
    public void setValue(Object value) {
        if (!(value instanceof Integer) || (value == null)) {
            setText("####");
        }
        setText(toHexString((Integer)value));
    }

    public static String toHexString(int i) {
        return (i < 0x10 ? "000" :
                i < 0x100 ? "00" :
                i < 0x1000 ? "0" : "")
                    + Integer.toHexString(i).toUpperCase();
    }
}
