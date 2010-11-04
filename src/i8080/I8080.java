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
package i8080;

/**
 *
 * @author Barend Scholtus
 */
public class I8080 {

    int B;
    int C;
    int D;
    int E;
    int H;
    int L;
    int A;
    int PC;
    int SP;
    boolean Carry;
    boolean Zero;
    boolean Sign;
    boolean Parity;
    boolean AuxCarry;
    boolean halted;
    boolean inte;
    private I8080Context ctx;

    public I8080(I8080Context env) {
        this.ctx = env;
    }

    public int reset() {
        PC = 0;
        inte = true;
        return 3;
    }

    public final int AF() {
        return (A << 8) | getConditionBits();
    }

    public final int BC() {
        return (B << 8) | C;
    }

    public final int DE() {
        return (D << 8) | E;
    }

    public final int HL() {
        return (H << 8) | L;
    }

    public final int getConditionBits() {
        return (Carry ? 0x01 : 0)
                | (0x02)
                | (Parity ? 0x04 : 0)
                | (AuxCarry ? 0x10 : 0)
                | (Zero ? 0x40 : 0)
                | (Sign ? 0x80 : 0);
    }

    private void setConditionBits(int t) {
        Carry = (t & 0x01) != 0;
        Parity = (t & 0x04) != 0;
        AuxCarry = (t & 0x10) != 0;
        Zero = (t & 0x40) != 0;
        Sign = (t & 0x80) != 0;
    }

//    private void writeByte(int addr, int data) {
//        ctx.writeByte(addr, data);
//    }
    public int interrupt(int opcode) {
        inte = false;
        return execute(opcode);
    }

    public int instructions(final int cycles) {
        if (halted) {
            return cycles;
        }

        int cyclesDone = 0;
        while (cyclesDone < cycles) {
            cyclesDone += execute(ctx.readByte(PC++));
        }

        return cyclesDone;
    }

    public int instruction() {
        if (halted) {
            return 1;
        }

        return execute(ctx.readByte(PC++));
    }

    private int execute(int opcode) {
        int t;

        switch (opcode) {
            case 0x00: // NOP
                return 4;
            case 0x01: // LXI B
                C = ctx.readByte(PC++);
                B = ctx.readByte(PC++);
                return 10;
            case 0x02: // STAX B
                ctx.writeByte(BC(), A);
                return 7;
            case 0x03: // INX B
                if (++C > 0xff) {
                    C = 0;
                    if (++B > 0xff) {
                        B = 0;
                    }
                }
                return 5;
            case 0x04: // INR B
                B = opINR(B);
                return 5;
            case 0x05: // DCR B
                B = opDCR(B);
                return 5;
            case 0x06: // MVI B
                B = ctx.readByte(PC++);
                return 7;
            case 0x07: // RLC
                Carry = (A & 0x80) != 0;
                A = ((A & 0x7f) << 1) | (Carry ? 0x1 : 0);
                return 4;
            case 0x08: // *NOP
                return 4;
            case 0x09: // DAD B
                opDAD(BC());
                return 10;
            case 0x0A: // LDAX B
                A = ctx.readByte(BC());
                return 7;
            case 0x0B: // DCX B
                if (--C < 0) {
                    C = 0xff;
                    if (--B < 0) {
                        B = 0xff;
                    }
                }
                return 5;
            case 0x0C: // INR C
                C = opINR(C);
                return 5;
            case 0x0D: // DCR C
                C = opDCR(C);
                return 5;
            case 0x0E: // MVI C
                C = ctx.readByte(PC++);
                return 7;
            case 0x0F: // RRC
                Carry = (A & 0x1) != 0;
                A = (A >>> 1) | (Carry ? 0x80 : 0);
                return 4;
            case 0x10: // *NOP
                return 4;
            case 0x11: // LXI D
                E = ctx.readByte(PC++);
                D = ctx.readByte(PC++);
                return 10;
            case 0x12: // STAX D
                ctx.writeByte(DE(), A);
                return 7;
            case 0x13: // INX D
                if (++E > 0xff) {
                    E = 0;
                    if (++D > 0xff) {
                        D = 0;
                    }
                }
                return 5;
            case 0x14: // INR D
                D = opINR(D);
                return 5;
            case 0x15: // DCR D
                D = opDCR(D);
                return 5;
            case 0x16: // MVI D
                D = ctx.readByte(PC++);
                return 7;
            case 0x17: // RAL
                t = Carry ? 1 : 0;
                Carry = (A & 0x80) != 0;
                A = ((A << 1) | t) & 0xff;
                return 4;
            case 0x18: // *NOP
                return 4;
            case 0x19: // DAD D
                opDAD(DE());
                return 10;
            case 0x1A: // LDAX D
                A = ctx.readByte(DE());
                return 7;
            case 0x1B: // DCX D
                if (--E < 0) {
                    E = 0xff;
                    if (--D < 0) {
                        D = 0xff;
                    }
                }
                return 5;
            case 0x1C: // INR E
                E = opINR(E);
                return 5;
            case 0x1D: // DCR E
                E = opDCR(E);
                return 5;
            case 0x1E: // MVI E
                E = ctx.readByte(PC++);
                return 7;
            case 0x1F: // RAR
                t = Carry ? 0x80 : 0;
                Carry = (A & 0x1) != 0;
                A = (A >>> 1) | t;
                return 4;
            case 0x20: // *NOP
                return 4;
            case 0x21: // LXI H
                L = ctx.readByte(PC++);
                H = ctx.readByte(PC++);
                return 10;
            case 0x22: // SHLD
                ctx.writeWord(ctx.readWord(PC), L, H);
                PC += 2;
                return 16;
            case 0x23: // INX H
                if (++L > 0xff) {
                    L = 0;
                    if (++H > 0xff) {
                        H = 0;
                    }
                }
                return 5;
            case 0x24: // INR H
                H = opINR(H);
                return 5;
            case 0x25: // DCR H
                H = opDCR(H);
                return 5;
            case 0x26: // MVI H
                H = ctx.readByte(PC++);
                return 7;
            case 0x27: // DAA
                opDAA();
                return 4;
            case 0x28: // *NOP
                return 4;
            case 0x29: // DAD H
                opDAD(HL());
                return 10;
            case 0x2A: // LHLD
                t = ctx.readWord(PC);
                L = ctx.readByte(t++);
                H = ctx.readByte(t);
                PC += 2;
                return 16;
            case 0x2B: // DCX H
                if (--L < 0) {
                    L = 0xff;
                    if (--H < 0) {
                        H = 0xff;
                    }
                }
                return 5;
            case 0x2C: // INR L
                L = opINR(L);
                return 5;
            case 0x2D: // DCR L
                L = opDCR(L);
                return 5;
            case 0x2E: // MVI L
                L = ctx.readByte(PC++);
                return 7;
            case 0x2F: // CMA
                A = ~A & 0xff;
                return 4;
            case 0x30: // *NOP
                return 4;
            case 0x31: // LXI SP
                SP = ctx.readWord(PC);
                PC += 2;
                ;
                return 10;
            case 0x32: // STA
                ctx.writeByte(ctx.readWord(PC), A);
                PC += 2;
                return 13;
            case 0x33: // INX SP
                ++SP;
                SP &= 0xffff;
                return 5;
            case 0x34: // INR M
                t = HL();
                ctx.writeByte(t, opINR(ctx.readByte(t)));
                return 10;
            case 0x35: // DCR M
                t = HL();
                ctx.writeByte(t, opDCR(ctx.readByte(t)));
                return 10;
            case 0x36: // MVI M
                ctx.writeByte(HL(), ctx.readByte(PC++));
                return 10;
            case 0x37: // STC
                Carry = true;
                return 4;
            case 0x38: // *NOP
                return 4;
            case 0x39: // DAD SP
                opDAD(SP);
                return 10;
            case 0x3A: // LDA
                A = ctx.readByte(ctx.readWord(PC));
                PC += 2;
                return 13;
            case 0x3B: // DCX SP
                --SP;
                SP &= 0xffff;
                return 5;
            case 0x3C: // INR A
                A = opINR(A);
                return 5;
            case 0x3D: // DCR A
                A = opDCR(A);
                return 5;
            case 0x3E: // MVI A
                A = ctx.readByte(PC++);
                return 7;
            case 0x3F: // CMC
                Carry = !Carry;
                return 4;
            case 0x40: // MOV
                /*B = B;*/
                return 5;
            case 0x41: // MOV
                B = C;
                return 5;
            case 0x42: // MOV
                B = D;
                return 5;
            case 0x43: // MOV
                B = E;
                return 5;
            case 0x44: // MOV
                B = H;
                return 5;
            case 0x45: // MOV
                B = L;
                return 5;
            case 0x46: // MOV
                B = ctx.readByte(HL());
                return 7;
            case 0x47: // MOV
                B = A;
                return 5;
            case 0x48: // MOV
                C = B;
                return 5;
            case 0x49: // MOV
                /*C = C;*/
                return 5;
            case 0x4A: // MOV
                C = D;
                return 5;
            case 0x4B: // MOV
                C = E;
                return 5;
            case 0x4C: // MOV
                C = H;
                return 5;
            case 0x4D: // MOV
                C = L;
                return 5;
            case 0x4E: // MOV
                C = ctx.readByte(HL());
                return 7;
            case 0x4F: // MOV
                C = A;
                return 5;
            case 0x50: // MOV
                D = B;
                return 5;
            case 0x51: // MOV
                D = C;
                return 5;
            case 0x52: // MOV
                /*D = D;*/
                return 5;
            case 0x53: // MOV
                D = E;
                return 5;
            case 0x54: // MOV
                D = H;
                return 5;
            case 0x55: // MOV
                D = L;
                return 5;
            case 0x56: // MOV
                D = ctx.readByte(HL());
                return 7;
            case 0x57: // MOV
                D = A;
                return 5;
            case 0x58: // MOV
                E = B;
                return 5;
            case 0x59: // MOV
                E = C;
                return 5;
            case 0x5A: // MOV
                E = D;
                return 5;
            case 0x5B: // MOV
                /*E = E;*/
                return 5;
            case 0x5C: // MOV
                E = H;
                return 5;
            case 0x5D: // MOV
                E = L;
                return 5;
            case 0x5E: // MOV
                E = ctx.readByte(HL());
                return 7;
            case 0x5F: // MOV
                E = A;
                return 5;
            case 0x60: // MOV
                H = B;
                return 5;
            case 0x61: // MOV
                H = C;
                return 5;
            case 0x62: // MOV
                H = D;
                return 5;
            case 0x63: // MOV
                H = E;
                return 5;
            case 0x64: // MOV
                /*H = H;*/
                return 5;
            case 0x65: // MOV
                H = L;
                return 5;
            case 0x66: // MOV
                H = ctx.readByte(HL());
                return 7;
            case 0x67: // MOV
                H = A;
                return 5;
            case 0x68: // MOV
                L = B;
                return 5;
            case 0x69: // MOV
                L = C;
                return 5;
            case 0x6A: // MOV
                L = D;
                return 5;
            case 0x6B: // MOV
                L = E;
                return 5;
            case 0x6C: // MOV
                L = H;
                return 5;
            case 0x6D: // MOV
                /*L = L;*/
                return 5;
            case 0x6E: // MOV
                L = ctx.readByte(HL());
                return 7;
            case 0x6F: // MOV
                L = A;
                return 5;
            case 0x70: // MOV
                ctx.writeByte(HL(), B);
                return 7;
            case 0x71: // MOV
                ctx.writeByte(HL(), C);
                return 7;
            case 0x72: // MOV
                ctx.writeByte(HL(), D);
                return 7;
            case 0x73: // MOV
                ctx.writeByte(HL(), E);
                return 7;
            case 0x74: // MOV
                ctx.writeByte(HL(), H);
                return 7;
            case 0x75: // MOV
                ctx.writeByte(HL(), L);
                return 7;
            case 0x76: // HLT
                halted = true;
                return 7;
            case 0x77: // MOV
                ctx.writeByte(HL(), A);
                return 7;
            case 0x78: // MOV
                A = B;
                return 5;
            case 0x79: // MOV
                A = C;
                return 5;
            case 0x7A: // MOV
                A = D;
                return 5;
            case 0x7B: // MOV
                A = E;
                return 5;
            case 0x7C: // MOV
                A = H;
                return 5;
            case 0x7D: // MOV
                A = L;
                return 5;
            case 0x7E: // MOV
                A = ctx.readByte(HL());
                return 7;
            case 0x7F: // MOV
                /*A = A;*/
                return 5;
            case 0x80: // ADD B
                opADD(B);
                return 4;
            case 0x81: // ADD C
                opADD(C);
                return 4;
            case 0x82: // ADD D
                opADD(D);
                return 4;
            case 0x83: // ADD E
                opADD(E);
                return 4;
            case 0x84: // ADD H
                opADD(H);
                return 4;
            case 0x85: // ADD L
                opADD(L);
                return 4;
            case 0x86: // ADD M
                opADD(ctx.readByte(HL()));
                return 7;
            case 0x87: // ADD A
                opADD(A);
                return 4;
            case 0x88: // ADC B
                opADC(B);
                return 4;
            case 0x89: // ADC C
                opADC(C);
                return 4;
            case 0x8A: // ADC D
                opADC(D);
                return 4;
            case 0x8B: // ADC E
                opADC(E);
                return 4;
            case 0x8C: // ADC H
                opADC(H);
                return 4;
            case 0x8D: // ADC L
                opADC(L);
                return 4;
            case 0x8E: // ADC M
                opADC(ctx.readByte(HL()));
                return 7;
            case 0x8F: // ADC A
                opADC(A);
                return 4;
            case 0x90: // SUB B
                opSUB(B);
                return 4;
            case 0x91: // SUB C
                opSUB(C);
                return 4;
            case 0x92: // SUB D
                opSUB(D);
                return 4;
            case 0x93: // SUB E
                opSUB(E);
                return 4;
            case 0x94: // SUB H
                opSUB(H);
                return 4;
            case 0x95: // SUB L
                opSUB(L);
                return 4;
            case 0x96: // SUB M
                opSUB(ctx.readByte(HL()));
                return 7;
            case 0x97: // SUB A
                opSUB(A);
                return 4;
            case 0x98: // SBB B
                opSBB(B);
                return 4;
            case 0x99: // SBB C
                opSBB(C);
                return 4;
            case 0x9A: // SBB D
                opSBB(D);
                return 4;
            case 0x9B: // SBB E
                opSBB(E);
                return 4;
            case 0x9C: // SBB H
                opSBB(H);
                return 4;
            case 0x9D: // SBB L
                opSBB(L);
                return 4;
            case 0x9E: // SBB M
                opSBB(ctx.readByte(HL()));
                return 7;
            case 0x9F: // SBB A
                opSBB(A);
                return 4;
            case 0xA0: // ANA B
                opANA(B);
                return 4;
            case 0xA1: // ANA C
                opANA(C);
                return 4;
            case 0xA2: // ANA D
                opANA(D);
                return 4;
            case 0xA3: // ANA E
                opANA(E);
                return 4;
            case 0xA4: // ANA H
                opANA(H);
                return 4;
            case 0xA5: // ANA L
                opANA(L);
                return 4;
            case 0xA6: // ANA M
                opANA(ctx.readByte(HL()));
                return 7;
            case 0xA7: // ANA A
                opANA(A);
                return 4;
            case 0xA8: // XRA B
                opXRA(B);
                return 4;
            case 0xA9: // XRA C
                opXRA(C);
                return 4;
            case 0xAA: // XRA D
                opXRA(D);
                return 4;
            case 0xAB: // XRA E
                opXRA(E);
                return 4;
            case 0xAC: // XRA H
                opXRA(H);
                return 4;
            case 0xAD: // XRA L
                opXRA(L);
                return 4;
            case 0xAE: // XRA M
                opXRA(ctx.readByte(HL()));
                return 7;
            case 0xAF: // XRA A
                opXRA(A);
                return 4;
            case 0xB0: // ORA B
                opORA(B);
                return 4;
            case 0xB1: // ORA C
                opORA(C);
                return 4;
            case 0xB2: // ORA D
                opORA(D);
                return 4;
            case 0xB3: // ORA E
                opORA(E);
                return 4;
            case 0xB4: // ORA H
                opORA(H);
                return 4;
            case 0xB5: // ORA L
                opORA(L);
                return 4;
            case 0xB6: // ORA M
                opORA(ctx.readByte(HL()));
                return 7;
            case 0xB7: // ORA A
                opORA(A);
                return 4;
            case 0xB8: // CMP B
                opCMP(B);
                return 4;
            case 0xB9: // CMP C
                opCMP(C);
                return 4;
            case 0xBA: // CMP D
                opCMP(D);
                return 4;
            case 0xBB: // CMP E
                opCMP(E);
                return 4;
            case 0xBC: // CMP H
                opCMP(H);
                return 4;
            case 0xBD: // CMP L
                opCMP(L);
                return 4;
            case 0xBE: // CMP M
                opCMP(ctx.readByte(HL()));
                return 7;
            case 0xBF: // CMP A
                opCMP(A);
                return 4;
            case 0xC0: // RNZ
                return opRET(!Zero);
            case 0xC1: // POP B
                C = ctx.readByte(SP++);
                B = ctx.readByte(SP++);
                return 10;
            case 0xC2: // JNZ
                return opJMP(!Zero);
            case 0xC3: // JMP
                PC = ctx.readWord(PC);
                return 10;
//                return opJMP(true);
            case 0xC4: // CNZ
                return opCALL(!Zero);
            case 0xC5: // PUSH B
                ctx.writeByte(--SP, B);
                ctx.writeByte(--SP, C);
                return 11;
            case 0xC6: // ADI
                opADD(ctx.readByte(PC++));
                return 7;
            case 0xC7: // RST 0
                opRST(0x00);
                return 11;
            case 0xC8: // RZ
                return opRET(Zero);
            case 0xC9: // RET
                PC = ctx.readWord(SP);
                SP += 2;
                return 10;
            case 0xCA: // JZ
                return opJMP(Zero);
            case 0xCB: // JMP
                PC = ctx.readWord(PC);
                return 10;
//                return opJMP(true);
            case 0xCC: // CZ
                return opCALL(Zero);
            case 0xCD: // CALL
                return opCALL(true);
            case 0xCE: // ACI
                opADC(ctx.readByte(PC++));
                return 7;
            case 0xCF: // RST 1
                opRST(0x08);
                return 11;
            case 0xD0: // RNC
                return opRET(!Carry);
            case 0xD1: // POP D
                E = ctx.readByte(SP++);
                D = ctx.readByte(SP++);
                return 10;
            case 0xD2: // JNC
                return opJMP(!Carry);
            case 0xD3: // OUT
                ctx.out(ctx.readByte(PC++), A);
                return 10;
            case 0xD4: // CNC
                return opCALL(!Carry);
            case 0xD5: // PUSH D
                ctx.writeByte(--SP, D);
                ctx.writeByte(--SP, E);
                return 11;
            case 0xD6: // SUI
                opSUB(ctx.readByte(PC++));
                return 7;
            case 0xD7: // RST 2
                opRST(0x10);
                return 11;
            case 0xD8: // RC
                return opRET(Carry);
            case 0xD9: // RET
                PC = ctx.readWord(SP);
                SP += 2;
                return 10;
            case 0xDA: // JC
                return opJMP(Carry);
            case 0xDB: // IN
                A = ctx.in(ctx.readByte(PC++));
                return 10;
            case 0xDC: // CC
                return opCALL(Carry);
            case 0xDD: // CALL
                return opCALL(true);
            case 0xDE: // SBI
                opSBB(ctx.readByte(PC++));
                return 7;
            case 0xDF: // RST 3
                opRST(0x18);
                return 11;
            case 0xE0: // RPO
                return opRET(!Parity);
            case 0xE1: // POP H
                L = ctx.readByte(SP++);
                H = ctx.readByte(SP++);
                return 10;
            case 0xE2: // JPO
                return opJMP(!Parity);
            case 0xE3: // XTHL
                opXTHL();
                return 18;
            case 0xE4: // CPO
                return opCALL(!Parity);
            case 0xE5: // PUSH H
                ctx.writeByte(--SP, H);
                ctx.writeByte(--SP, L);
                return 11;
            case 0xE6: // ANI
                opANA(ctx.readByte(PC++));
                return 7;
            case 0xE7: // RST 4
                opRST(0x20);
                return 11;
            case 0xE8: // RPE
                return opRET(Parity);
            case 0xE9: // PCHL
                PC = (H << 8) | L;
                return 5;
            case 0xEA: // JPE
                return opJMP(Parity);
            case 0xEB: // XCHG
                opXCHG();
                return 4;
            case 0xEC: // CPE
                return opCALL(Parity);
            case 0xED: // CALL
                return opCALL(true);
            case 0xEE: // XRI
                opXRA(ctx.readByte(PC++));
                return 7;
            case 0xEF: // RST 5
                opRST(0x28);
                return 11;
            case 0xF0: // RP
                return opRET(!Sign);
            case 0xF1: // POP PSW
                setConditionBits(ctx.readByte(SP++));
                A = ctx.readByte(SP++);
                return 10;
            case 0xF2: // JP
                return opJMP(!Sign);
            case 0xF3:
                inte = false;
                return 4;
            case 0xF4: // CP
                return opCALL(!Sign);
            case 0xF5: // PUSH PSW
                ctx.writeByte(--SP, A);
                ctx.writeByte(--SP, getConditionBits());
                return 11;
            case 0xF6: // ORI
                opORA(ctx.readByte(PC++));
                return 7;
            case 0xF7: // RST 6
                opRST(0x30);
                return 11;
            case 0xF8: // RM
                return opRET(Sign);
            case 0xF9: // SPHL
                SP = (H << 8) | L;
                return 5;
            case 0xFA: // JM
                return opJMP(Sign);
            case 0xFB:
                inte = true;
                return 4;
            case 0xFC: // CM
                return opCALL(Sign);
            case 0xFD: // CALL
                return opCALL(true);
            case 0xFE: // CPI
                opCMP(ctx.readByte(PC++));
                return 7;
            case 0xFF: // RST 7
                opRST(0x38);
                return 11;
            default:
                throw new RuntimeException("opcode " + opcode + " does not exist");
        }
    }

    private int opJMP(final boolean cond) {
        if (cond) {
            PC = ctx.readWord(PC);
            return 10;
        } else {
            PC += 2;
            return 3;
        }
    }

    private int opCALL(final boolean cond) {
        if (cond) {
//            ctx.writeByte(--SP, (PC + 2) >>> 8);
//            ctx.writeByte(--SP, (PC + 2) & 0xff);
            SP -= 2;
            ctx.writeWord(SP, (PC + 2) & 0xff, (PC + 2) >>> 8);
            PC = ctx.readWord(PC);
            return 17;
        } else {
            PC += 2;
            return 11;
        }
    }

    private void opRST(final int exp) {
        ctx.writeByte(--SP, (PC >>> 8));
        ctx.writeByte(--SP, (PC & 0xff));
        PC = exp;
    }

    private int opRET(final boolean cond) {
        if (cond) {
            PC = ctx.readWord(SP);
            SP += 2;
            return 11;
        } else {
            return 5;
        }
    }

    private void opXTHL() {
        int t = L;
        L = ctx.readByte(SP);
        ctx.writeByte(SP, t);
        t = H;
        H = ctx.readByte(SP + 1);
        ctx.writeByte(SP + 1, t);
    }

    private void opXCHG() {
        int t = H;
        H = D;
        D = t;
        t = L;
        L = E;
        E = t;
    }

//    private void setSZPFrom(final int val) {
//        Sign = (val & 0x80) != 0;
//        Zero = (val == 0);
//        Parity = evenParity[val];
////        Parity = evenParity(val);
//    }

    private void opDAD(int other) {
        other += HL();
        Carry = (other & 0x10000) != 0;
        H = (other >>> 8) & 0xff;
        L = other & 0xff;
    }

    private void opDAA() {
        if ((A & 0x0f) > 9 || AuxCarry) {
            A += 6;
            AuxCarry = true;
        } else {
            AuxCarry = false;
        }
        if ((A >>> 4) > 9 || Carry) {
            A &= 0x0f;
            Carry = true;
        }
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
    }

    private int opINR(int reg) {
        ++reg;
        reg &= 0xff;
        AuxCarry = (reg & 0xf) == 0;
//        setSZPFrom(reg);
        Sign = (reg & 0x80) != 0;
        Zero = (reg == 0);
        Parity = evenParity[reg];
        return reg;
    }

    private int opDCR(int reg) {
        --reg;
        reg &= 0xff;
        AuxCarry = (reg & 0xf) == 0xf;
        Sign = (reg & 0x80) != 0;
        Zero = (reg == 0);
        Parity = evenParity[reg];
//        setSZPFrom(reg);
        return reg;
    }

    private void opADD(final int reg) {
        int t = A + reg;
        Carry = (t & 0x100) != 0;
        AuxCarry = ((A ^ t ^ reg) & 0x10) != 0;
        A = t & 0xff;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
    }

    private void opADC(final int reg) {
        int t = A + reg + (Carry ? 1 : 0);
        Carry = (t & 0x100) != 0;
        AuxCarry = ((A ^ t ^ reg) & 0x10) != 0;
        A = t & 0xff;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
    }

    private void opSUB(final int reg) {
        int t = A - reg;
        Carry = (t & 0x100) != 0;
        AuxCarry = ((A ^ t ^ reg) & 0x10) != 0;
        A = t & 0xff;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
    }

    private void opSBB(final int reg) {
        int t = A - reg - (Carry ? 1 : 0);
        Carry = (t & 0x100) != 0;
        AuxCarry = ((A ^ t ^ reg) & 0x10) != 0;
        A = t & 0xff;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
    }

    private void opANA(final int reg) {
        A &= reg;
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
//        Parity = evenParity(A);
//        setSZPFrom(A);
        Carry = false;
    }

    private void opXRA(final int reg) {
        A ^= reg;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
        Carry = false;
    }

    private void opORA(final int reg) {
        A |= reg;
//        setSZPFrom(A);
        Sign = (A & 0x80) != 0;
        Zero = (A == 0);
        Parity = evenParity[A];
        Carry = false;
    }

    private void opCMP(final int reg) {
        int t = A - reg;
        Carry = (t & 0x100) != 0;
        AuxCarry = ((A ^ t ^ reg) & 0x10) != 0;
        t &= 0xff;
        Sign = (t & 0x80) != 0;
        Zero = (t == 0);
        Parity = evenParity[t];
    }

    public final int PC() {
        return PC;
    }

    public final int SP() {
        return SP;
    }

    public final boolean isCarry() {
        return Carry;
    }

    public final boolean isParity() {
        return Parity;
    }

    public final boolean isAuxCarry() {
        return AuxCarry;
    }

    public final boolean isZero() {
        return Zero;
    }

    public final boolean isSign() {
        return Sign;
    }

    /**
     * Returns whether val has even parity or not.
     * Based on http://graphics.stanford.edu/~seander/bithacks.html
     * @param val the value for which to determine whether it has even parity
     * @return true if val has even parity
     */
//    public static boolean evenParity(int val) {
//        return (0x6996 >> ((val ^ (val >> 4)) & 0xF) & 1) == 0;
//    }

    public static final boolean[] evenParity = {
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        false, true, true, false, true, false, false, true, true, false, false, true, false, true, true, false,
        true, false, false, true, false, true, true, false, false, true, true, false, true, false, false, true
    };
}
