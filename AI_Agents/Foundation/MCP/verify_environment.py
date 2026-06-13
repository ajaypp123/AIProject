#!/usr/bin/env python3
"""Verify MCP + LangGraph environment setup"""

import os
import sys
import importlib

def check_package(package_name, import_name=None):
    """Check if a package is installed and can be imported"""
    if import_name is None:
        import_name = package_name

    try:
        module = importlib.import_module(import_name)
        version = getattr(module, "__version__", "unknown")
        print(f"  ✅ {package_name}: {version}")
        return True
    except ImportError:
        print(f"  ❌ {package_name}: Not installed")
        return False

def check_env_var(var_name):
    """Check if an environment variable is set"""
    value = os.getenv(var_name)
    if value:
        # Mask sensitive values
        if "KEY" in var_name or "TOKEN" in var_name:
            masked = value[:4] + "..." + value[-4:] if len(value) > 8 else "***"
            print(f"  ✅ {var_name}: {masked}")
        else:
            print(f"  ✅ {var_name}: {value}")
        return True
    else:
        print(f"  ⚠️ {var_name}: Not set")
        return False

def main():
    print("=" * 60)
    print("🔍 MCP + LangGraph Environment Verification")
    print("=" * 60)

    all_good = True

    # Check Python version
    print("\n📦 Python Version:")
    python_version = sys.version.split()[0]
    if sys.version_info >= (3, 8):
        print(f"  ✅ Python: {python_version}")
    else:
        print(f"  ❌ Python: {python_version} (Need 3.8+)")
        all_good = False

    # Check required packages
    print("\n📦 Required Packages:")
    packages = [
        ("langgraph", "langgraph"),
        ("langchain", "langchain"),
        ("langchain-openai", "langchain_openai"),
        ("langchain-community", "langchain_community"),
    ]

    for package, import_name in packages:
        if not check_package(package, import_name):
            all_good = False

    # Check optional MCP packages (will be available in future)
    print("\n📦 Optional MCP Packages:")
    optional_packages = [
        ("fastmcp", "mcp"),  # Future package
        ("langchain-mcp", "langchain_mcp"),  # Future package
    ]

    for package, import_name in optional_packages:
        try:
            check_package(package, import_name)
        except:
            print(f"  ℹ️ {package}: Not available yet (using mock implementation)")

    # Check environment variables
    print("\n🔐 Environment Variables:")
    env_vars = [
        "OPENAI_API_BASE",
        "OPENAI_API_KEY",
        "OPENAI_MODEL"
    ]

    for var in env_vars:
        if not check_env_var(var):
            if var != "OPENAI_MODEL":  # Model is optional
                all_good = False

    # Check directories
    print("\n📁 Directory Structure:")
    directories = [
        "/root/code",
        "/root/code/mcp_servers",
        "/root/markers",
        "/root/mcp-project"
    ]

    for dir_path in directories:
        if os.path.exists(dir_path):
            print(f"  ✅ {dir_path}")
        else:
            print(f"  ❌ {dir_path}: Not found")
            all_good = False

    # Check task files
    print("\n📄 Task Files:")
    task_files = [
        "/root/code/task_1_mcp_basics.py",
        "/root/code/task_2_mcp_langgraph.py",
        "/root/code/task_3_multi_servers.py"
    ]

    for file_path in task_files:
        if os.path.exists(file_path):
            print(f"  ✅ {os.path.basename(file_path)}")
        else:
            print(f"  ❌ {os.path.basename(file_path)}: Not found")
            all_good = False

    # Final status
    print("\n" + "=" * 60)
    if all_good:
        print("✅ Environment is ready for MCP + LangGraph lab!")
        print("\n🚀 Next Steps:")
        print("  1. Start with Task 1: python3 /root/code/task_1_mcp_basics.py")
        print("  2. Continue to Task 2: python3 /root/code/task_2_mcp_langgraph.py")
        print("  3. Complete Task 3: python3 /root/code/task_3_multi_servers.py")

        # Create marker file
        os.makedirs("/root/markers", exist_ok=True)
        with open("/root/markers/environment_verified.txt", "w") as f:
            f.write("ENVIRONMENT_VERIFIED")
        print("\n✅ Marker created: /root/markers/environment_verified.txt")
    else:
        print("❌ Environment setup incomplete. Please install missing packages:")
        print("\n  pip install langgraph langchain langchain-openai")
        sys.exit(1)

    print("=" * 60)

if __name__ == "__main__":
    main()