import os
import re
import shutil

root_dir = r"C:\Users\tilak\.gemini\antigravity\worktrees\oenexa\fix_errors"

replacements = [
    (r"qbc-qubitscoin/qubitscoin", "qbc-qubitscoin/oenexa-node"),
    (r"qbc-qubitscoin/qbc-node", "qbc-qubitscoin/oenexa-node"),
    (r"QubitsCoin", "OENEXA"),
    (r"qubitscoin\.io", "oenexa.io"),
    (r"qubitscoin", "oenexa"),
    (r"QubitCoin", "OENEXA"),
    (r"Qubits", "Oenexa"),
    (r"qubits", "oenexa"),
    (r"Qubit", "Oenexa"),
    (r"qubit", "oenexa"),
    (r"qbc-node", "oenexa-node"),
    (r"qbc_node", "oenexa_node"),
    (r"QBC_PASSWORD", "OEN_PASSWORD"),
    (r"qbc-seed", "oenexa-seed"),
    (r"qbc-testnet", "oenexa-testnet"),
    (r"qbc_", "oen_"),
    (r"QBC_", "OEN_"),
    (r"QBC", "OEN"),
    (r"\bqbc\b", "oen"),
    (r"\bQbc\b", "Oen"),
    (r"\bOneOEN\s+uint64\s*=\s*OneOEN\b", ""),
]

def replace_in_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception:
        return False
    
    new_content = content
    for pattern, repl in replacements:
        new_content = re.sub(pattern, repl, new_content)
        
    if new_content != content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        return True
    return False

def rename_paths(dir_path):
    for root, dirs, files in os.walk(dir_path, topdown=False):
        if '.git' in root or 'node_modules' in root:
            continue
            
        for name in files + dirs:
            if name == 'mass_replace.py':
                continue
                
            new_name = name
            for pattern, repl in replacements:
                new_name = re.sub(pattern, repl, new_name)
                
            if new_name != name:
                old_path = os.path.join(root, name)
                new_path = os.path.join(root, new_name)
                try:
                    os.rename(old_path, new_path)
                    print(f"Renamed: {old_path} -> {new_path}")
                except FileExistsError:
                    print(f"File {new_path} already exists. Deleting {old_path}")
                    try:
                        if os.path.isdir(old_path):
                            shutil.rmtree(old_path)
                        else:
                            os.remove(old_path)
                    except Exception as e:
                        print(f"Failed to remove {old_path}: {e}")

def replace_content(dir_path):
    for root, dirs, files in os.walk(dir_path):
        if '.git' in root or 'node_modules' in root:
            continue
        for name in files:
            if name == 'mass_replace.py':
                continue
            filepath = os.path.join(root, name)
            if replace_in_file(filepath):
                print(f"Updated content: {filepath}")

rename_paths(root_dir)
replace_content(root_dir)
print("Mass replace completed.")
