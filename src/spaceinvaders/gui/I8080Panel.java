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
    private int[] initialColumnsWidth = {
        50, 40, 60, 90
    };
    private volatile Thread runToCursorThread;
    private I8080Context ctx;

    /**
     *
     * @param ctx
     */
    public I8080Panel(I8080Context ctx) {
        this.ctx = ctx;
        createComponents();
        btnStep.addActionListener(this);
        btnRunTC.addActionListener(this);
        btnStop.addActionListener(this);
        btnReset.addActionListener(this);
        tableModel.setIsRowAnOpCode(ctx.getCpu().PC());
        updateDisplay();
    }

    /**
     *
     */
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

        setLayout(new BorderLayout());
        add(topPanel, BorderLayout.NORTH);
        add(scrollPane);

        table.getColumnModel().getColumn(0).setCellRenderer(
                new DefaultTableCellRenderer() {

                    @Override
                    public void setValue(Object value) {
                        if (!(value instanceof Integer) || (value == null)) {
                            setText("????");
                        }
                        setText(Util.toHexString((Integer) value, 4));
                    }
                });
        table.getColumnModel().getColumn(1).setCellRenderer(
                new DefaultTableCellRenderer() {

                    @Override
                    public void setValue(Object value) {
                        if (!(value instanceof Integer) || (value == null)) {
                            setText("??");
                        }
                        setText(Util.toHexString((Integer) value, 2));
                    }
                });
        setInitialColumnWidths();
        table.setColumnSelectionAllowed(false);
        table.setRowSelectionAllowed(true);
        table.setSelectionMode(ListSelectionModel.SINGLE_SELECTION);
        table.setShowHorizontalLines(false);
        table.setAutoResizeMode(JTable.AUTO_RESIZE_OFF);
    }

    /**
     *
     */
    private void updateDisplay() {
        I8080 cpu = ctx.getCpu();

        fldPC.setText(Util.toHexString(cpu.PC(), 4));
        fldSP.setText(Util.toHexString(cpu.SP(), 4));
        fldAF.setText(Util.toHexString(cpu.AF(), 4));
        fldBC.setText(Util.toHexString(cpu.BC(), 4));
        fldDE.setText(Util.toHexString(cpu.DE(), 4));
        fldHL.setText(Util.toHexString(cpu.HL(), 4));
        fldFlags.setText((cpu.isCarry() ? "1" : "0")
                + (cpu.isParity() ? "1" : "0")
                + (cpu.isAuxCarry() ? "1" : "0")
                + (cpu.isZero() ? "1" : "0")
                + (cpu.isSign() ? "1" : "0"));
        fldCycles.setText(Integer.toString(cycles));

        ListSelectionModel lsm = table.getSelectionModel();
        lsm.setSelectionInterval(cpu.PC(), cpu.PC());
        table.scrollRectToVisible(table.getCellRect(cpu.PC(), 0, true));
    }

    /**
     *
     */
    private void reset() {
        ctx.getCpu().reset();
        cycles = 0;
        tableModel.setIsRowAnOpCode(ctx.getCpu().PC());
        updateDisplay();
    }

    /**
     *
     * @param e
     */
    public void actionPerformed(ActionEvent e) {
        Object src = e.getSource();

        if (src == btnStep) {
            I8080 cpu = ctx.getCpu();
            cycles += cpu.instruction();
            tableModel.setIsRowAnOpCode(cpu.PC());
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

    /**
     *
     * @param toPC
     */
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

    /**
     *
     */
    private void stopRunToCursor() {
        runToCursorThread = null;
    }

    /**
     *
     */
    private void setInitialColumnWidths() {
        int tot = 0;
        for (int i = 0; i < table.getColumnCount() - 1; i++) {
            table.getColumnModel().getColumn(i).setPreferredWidth(initialColumnsWidth[i]);
            tot += initialColumnsWidth[i];
        }
        table.getColumnModel().getColumn(table.getColumnCount() - 1).
                setPreferredWidth(table.getPreferredScrollableViewportSize().width - tot);
    }
}
